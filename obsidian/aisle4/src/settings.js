import { Notice, PluginSettingTab, Setting, requestUrl } from "obsidian";

// ─────────────────────────────────────────────────────────────────────────────
// Default settings
// ─────────────────────────────────────────────────────────────────────────────

export const DEFAULT_SETTINGS = {
  apiBaseUrl: "https://groceries.taiidani.com",
  username: "",
  password: "",
  token: "",
  tokenExpiresAt: "",
};

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

    // ── Credentials ───────────────────────────────────────────────────────

    new Setting(containerEl).setName("Username").addText((text) =>
      text
        .setPlaceholder("username")
        .setValue(this.plugin.settings.username)
        .onChange(async (value) => {
          this.plugin.settings.username = value.trim();
          await this.plugin.saveSettings();
        }),
    );

    new Setting(containerEl).setName("Password").addText((text) => {
      text
        .setPlaceholder("password")
        .setValue(this.plugin.settings.password)
        .onChange(async (value) => {
          this.plugin.settings.password = value;
          await this.plugin.saveSettings();
        });
      // Obsidian's addText() doesn't expose a password type natively
      text.inputEl.type = "password";
    });

    // ── Connection status + Connect button (same row) ─────────────────────

    const connectionSetting = new Setting(containerEl)
      .setName("Connection")
      .addButton((button) => {
        button
          .setButtonText("Connect")
          .setCta()
          .onClick(async () => {
            button.setButtonText("Connecting…");
            button.setDisabled(true);
            try {
              await this.fetchToken();
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

  // POSTs credentials to the API and stores the returned token in settings.
  // Uses requestUrl() instead of fetch() to bypass Electron's CORS restrictions.
  async fetchToken() {
    const { apiBaseUrl, username, password } = this.plugin.settings;
    if (!apiBaseUrl) throw new Error("API base URL is required.");
    if (!username) throw new Error("Username is required.");
    if (!password) throw new Error("Password is required.");

    const response = await requestUrl({
      url: `${apiBaseUrl}/api/v1/auth/login`,
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
      throw: false,
    });

    if (response.status < 200 || response.status >= 300) {
      const body = response.json || {};
      throw new Error(body.error || `Server returned ${response.status}`);
    }

    const data = response.json;
    this.plugin.settings.token = data.token;
    this.plugin.settings.tokenExpiresAt = data.expires_at;
    await this.plugin.saveSettings();
  }
}
