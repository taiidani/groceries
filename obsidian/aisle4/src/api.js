import { requestUrl } from "obsidian";

// ─────────────────────────────────────────────────────────────────────────────
// API helpers
// ─────────────────────────────────────────────────────────────────────────────

/**
 * Fetches the full item catalog and checks each item's name against it.
 *
 * Returns an array of result objects in the same order as `items`:
 *   status:           'known' | 'on-list' | 'new'
 *   canonicalName:    API's stored name (may differ in case), or null for 'new'
 *   listItemId:       catalog item ID needed for PUT, or null when not on list
 *   existingQuantity: current quantity string on the list, or null when not on list
 *
 * Throws if token is missing or if the catalog request fails.
 */
export async function verifyItems(items, settings) {
  if (!settings.token) {
    throw new Error("Not connected. Open Settings → Aisle4 and click Connect.");
  }

  const response = await requestUrl({
    url: `${settings.apiBaseUrl}/api/v1/items`,
    method: "GET",
    headers: {
      Authorization: `Bearer ${settings.token}`,
    },
    throw: false,
  });

  if (response.status !== 200) {
    const body = response.json || {};
    throw new Error(body.error || `Server returned ${response.status}`);
  }

  const catalog = response.json; // array of { id, name, list, ... }

  return items.map((item) => {
    const nameLower = item.name.trim().toLowerCase();
    const match = catalog.find((ci) => ci.name.toLowerCase() === nameLower);
    if (!match) {
      return {
        status: "new",
        canonicalName: null,
        listItemId: null,
        existingQuantity: null,
      };
    }
    const onList = match.list !== null;
    return {
      status: onList ? "on-list" : "known",
      canonicalName: match.name,
      listItemId: onList ? match.id : null,
      existingQuantity: onList ? match.list.quantity : null,
    };
  });
}

// Joins two free-form quantity strings with " + ". If one side is blank the
// other is returned as-is, avoiding a spurious " + " when a quantity is empty.
function appendQuantities(existing, addition) {
  const e = (existing || "").trim();
  const a = (addition || "").trim();
  if (!e) return a;
  if (!a) return e;
  return `${e} + ${a}`;
}

/**
 * Adds each item in `items` to the grocery list sequentially.
 *
 * Items whose `listItemId` is set (verified as already on the list) are
 * updated via PUT with an appended quantity rather than POSTed again.
 *
 * Returns { added: number, appended: number, alreadyOnList: number, errors: string[] }.
 * Throws if token is missing.
 */
export async function addToGroceryList(items, settings) {
  if (!settings.token) {
    throw new Error("Not connected. Open Settings → Aisle4 and click Connect.");
  }

  let added = 0;
  let appended = 0;
  let alreadyOnList = 0;
  const errors = [];

  for (const item of items) {
    if (item.listItemId != null) {
      // Item is already on the list — append to the existing quantity.
      const response = await requestUrl({
        url: `${settings.apiBaseUrl}/api/v1/list/items/${item.listItemId}`,
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${settings.token}`,
        },
        body: JSON.stringify({
          quantity: appendQuantities(item.existingQuantity, item.quantity),
        }),
        throw: false,
      });
      if (response.status === 200) {
        appended++;
      } else {
        const body = response.json || {};
        errors.push(`${item.name}: ${body.error || `HTTP ${response.status}`}`);
      }
    } else {
      const response = await requestUrl({
        url: `${settings.apiBaseUrl}/api/v1/list/items`,
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${settings.token}`,
        },
        body: JSON.stringify({ name: item.name, quantity: item.quantity }),
        throw: false,
      });
      if (response.status === 201) {
        added++;
      } else if (response.status === 409) {
        alreadyOnList++;
      } else {
        const body = response.json || {};
        errors.push(`${item.name}: ${body.error || `HTTP ${response.status}`}`);
      }
    }
  }

  return { added, appended, alreadyOnList, errors };
}
