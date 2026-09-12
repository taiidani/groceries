import { Notice, PluginSettingTab, Setting, requestUrl } from "obsidian";

// ─────────────────────────────────────────────────────────────────────────────
// Authelia OIDC configuration
// ─────────────────────────────────────────────────────────────────────────────

// Aisle4 is registered in Authelia as a public client using the OAuth 2.0
// Device Authorization Grant (RFC 8628): the plugin has no redirect target
// to receive an authorization code (and must work on Obsidian mobile, which
// can't run a local HTTP listener), so it instead shows the user a link +
// code to approve in a normal browser and polls for the result. No client
// secret is used or needed for this flow.
const AUTHELIA_BASE_URL = "https://auth.taiidani.com";
const OIDC_CLIENT_ID = "groceries-obsidian";
const OIDC_SCOPE = "openid profile email groups";
const DEFAULT_POLL_INTERVAL_SECONDS = 5;

// ─────────────────────────────────────────────────────────────────────────────
// Default settings
// ─────────────────────────────────────────────────────────────────────────────

export const DEFAULT_SETTINGS = {
  apiBaseUrl: "https://groceries.taiidani.com",
  token: "",
  tokenExpiresAt: "",
};

// ─────────────────────────────────────────────────────────────────────────────
// Device Authorization Grant helpers
// ─────────────────────────────────────────────────────────────────────────────

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// Starts the device flow, returning the device/user codes and verification
// link the caller should display to the user.
async function requestDeviceAuthorization() {
  const response = await requestUrl({
    url: `${AUTHELIA_BASE_URL}/api/oidc/device-authorization`,
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      client_id: OIDC_CLIENT_ID,
      scope: OIDC_SCOPE,
    }).toString(),
    throw: false,
  });

  if (response.status < 200 || response.status >= 300) {
    const body = response.json || {};
    throw new Error(
      body.error_description ||
        body.error ||
        `Server returned ${response.status}`,
    );
  }

  return response.json; // { device_code, user_code, verification_uri, verification_uri_complete, expires_in, interval }
}

// Polls Authelia's token endpoint until the user approves the device code (or
// it's denied/expires), resolving with the granted access token.
async function pollForDeviceToken(device) {
  const intervalMs =
    (device.interval || DEFAULT_POLL_INTERVAL_SECONDS) * 1000;
  const deadline = Date.now() + (device.expires_in || 600) * 1000;

  while (Date.now() < deadline) {
    await sleep(intervalMs);

    const response = await requestUrl({
      url: `${AUTHELIA_BASE_URL}/api/oidc/token`,
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "urn:ietf:params:oauth:grant-type:device_code",
        device_code: device.device_code,
        client_id: OIDC_CLIENT_ID,
      }).toString(),
      throw: false,
    });

    const body = response.json || {};
    if (response.status >= 200 && response.status < 300) {
      return body.access_token;
    }

    switch (body.error) {
      case "authorization_pending":
        continue;
      case "slow_down":
        // Authelia is asking us to back off further; a fixed extra wait is
        // simpler than tracking a growing interval and is fine for this UX.
        await sleep(intervalMs);
        continue;
      case "access_denied":
        throw new Error("Login was denied.");
      case "expired_token":
        throw new Error("The login code expired. Please try again.");
      default:
        throw new Error(
          body.error_description ||
            body.error ||
            `Server returned ${response.status}`,
        );
    }
  }

  throw new Error("Timed out waiting for login approval.");
}

// Exchanges an Authelia access token for this app's own long-lived API token.
async function exchangeAccessToken(apiBaseUrl, accessToken) {
  const response = await requestUrl({
    url: `${apiBaseUrl}/api/v1/auth/login`,
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ access_token: accessToken }),
    throw: false,
  });

  if (response.status < 200 || response.status >= 300) {
    const body = response.json || {};
    throw new Error(body.error || `Server returned ${response.status}`);
  }

  return response.json; // { token, expires_at }
}

// ─────────────────────────────────────────────────────────────────────────────
// Settings tab
// ─────────────────────────────────────────────────────────────────────────────

export class Aisle4SettingTab extends PluginSettingTab {
  constructor(app, plugin) {
    super(app, plugin);
    this.plugin = plugin;
  }

  display() {
    const { containerEl } = this;
    containerEl.empty();

    // ── Server URL ────────────────────────────────────────────────────────

    new Setting(containerEl)
      .setName("API base URL")
      .setDesc(
        "The base URL of your Aisle4 server (e.g. https://groceries.taiidani.com).",
      )
      .addText((text) =>
        text
          .setPlaceholder("https://groceries.taiidani.com")
          .setValue(this.plugin.settings.apiBaseUrl)
          .onChange(async (value) => {
            this.plugin.settings.apiBaseUrl = value.trim();
            await this.plugin.saveSettings();
          }),
      );

    // ── Connection status + Connect button (same row) ─────────────────────

    const connectionSetting = new Setting(containerEl)
      .setName("Connection")
      .addButton((button) => {
        button
          .setButtonText("Connect")
          .setCta()
          .onClick(async () => {
            button.setDisabled(true);
            try {
              await this.connectViaDeviceFlow(connectionSetting.descEl, button);
              this.renderConnectionStatus(connectionSetting.descEl);
              new Notice("Aisle4: Connected successfully!");
            } catch (err) {
              connectionSetting.descEl.empty();
              connectionSetting.descEl.createSpan({
                text: "Failed: " + err.message,
                cls: "aisle4-status-error",
              });
              new Notice("Aisle4: " + err.message, 6000);
            } finally {
              button.setButtonText("Connect");
              button.setDisabled(false);
            }
          });
      });

    this.renderConnectionStatus(connectionSetting.descEl);
  }

  // Renders the current connection state into a container element.
  renderConnectionStatus(el) {
    el.empty();
    const { token, tokenExpiresAt } = this.plugin.settings;

    if (!token) {
      el.createSpan({
        text: "Not connected.",
        cls: "aisle4-status-disconnected",
      });
      return;
    }

    if (tokenExpiresAt) {
      const expiry = new Date(tokenExpiresAt);
      if (expiry <= new Date()) {
        el.createSpan({
          text: "Token expired — please reconnect.",
          cls: "aisle4-status-error",
        });
        return;
      }
      el.createSpan({
        text: `Connected. Token expires ${expiry.toLocaleDateString()}.`,
        cls: "aisle4-status-ok",
      });
      return;
    }

    el.createSpan({ text: "Connected.", cls: "aisle4-status-ok" });
  }

  // Runs the full device authorization flow: requests a device/user code,
  // displays it for the user to approve in a browser, polls until approved,
  // then exchanges the resulting Authelia access token for our own API token.
  async connectViaDeviceFlow(descEl, button) {
    const { apiBaseUrl } = this.plugin.settings;
    if (!apiBaseUrl) throw new Error("API base URL is required.");

    button.setButtonText("Requesting code…");
    const device = await requestDeviceAuthorization();

    const link = device.verification_uri_complete || device.verification_uri;
    descEl.empty();
    descEl.createSpan({ text: "Go to " });
    const linkEl = descEl.createEl("a", { text: link, href: link });
    linkEl.target = "_blank";
    descEl.createEl("br");
    descEl.createSpan({ text: "and approve code " });
    descEl.createEl("strong", { text: device.user_code });

    button.setButtonText("Waiting for approval…");
    const accessToken = await pollForDeviceToken(device);

    button.setButtonText("Finishing up…");
    const { token, expires_at } = await exchangeAccessToken(
      apiBaseUrl,
      accessToken,
    );

    this.plugin.settings.token = token;
    this.plugin.settings.tokenExpiresAt = expires_at;
    await this.plugin.saveSettings();
  }
}
