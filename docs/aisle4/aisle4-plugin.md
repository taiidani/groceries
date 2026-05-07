# Aisle4 — Obsidian Plugin

**Last updated:** 2026-05-06
**Status:** Complete

---

## Overview

Aisle4 is a personal Obsidian plugin that bridges recipe notes and the Aisle4
grocery list API. With the shopping-cart ribbon icon, you can scrape ingredients
from a recipe note open in Recipe View, review and correct them in a modal
dialog, verify them against the live item catalog, and add them to your grocery
list — all without leaving Obsidian.

The plugin lives in the groceries monorepo (not in the vault) and is deployed
by a `mise` build task. It has no npm dependencies; esbuild is pulled in via a
Go tool directive.

---

## Prerequisites

### Obsidian plugins

Aisle4 hard-depends on
**[lachholden/obsidian-recipe-view](https://github.com/lachholden/obsidian-recipe-view)**
being installed and enabled in the vault. The recipe note must be *open in Recipe
View mode* before clicking the cart icon — Aisle4 reads the live rendered DOM,
not the raw markdown file.

If the active note does not have a Recipe View leaf open, clicking the ribbon
icon shows a friendly error and does nothing else.

### API connection

Aisle4 talks to `https://groceries.taiidani.com` by default (configurable in
Settings). Before using the plugin you need to connect it to the API:

1. Open **Settings → Aisle4**.
2. Enter your username and password.
3. Click **Connect**. The plugin calls `POST /api/v1/auth/login`, stores the
   returned 30-day bearer token, and displays the expiry date in green when
   successful.

The token survives Obsidian restarts because it is saved in the plugin's data
store alongside the other settings.

---

## User Flow

The complete workflow from a user's perspective:

1. Open a recipe note and switch it to Recipe View mode (via the view header or
   the command palette).
2. Optionally scale the serving count using recipe-view's controls — Aisle4 will
   capture the *scaled* quantities.
3. Click the shopping-cart icon in the left ribbon.
4. The **Add to Grocery List** modal opens, showing every ingredient with a
   checkbox, a quantity field, and a name field. Ingredients you had already
   crossed off in Recipe View arrive pre-unchecked. A faint hint line below each
   row shows the raw text scraped from the DOM, useful as a reference when
   correcting the parsed fields.
5. Edit any names or quantities you want to fix, then click **Verify**. Aisle4
   fetches the full item catalog in a single API call and checks each ingredient
   name:
   - **Known** (green) — the item exists in the catalog and is not currently on
     the shopping list. It will be added as-is.
   - **On list** (red) — the item is already on the shopping list. Submitting
     will *append* your quantity to the existing one rather than overwriting it
     or skipping it.
   - **New** (muted) — no matching item was found. Submitting will create a new
     uncategorized item automatically.

   If the only difference between what you typed and what the API has stored is
   letter case, Aisle4 silently corrects your entry to match the catalog's
   capitalization, since the API is case-sensitive.

6. After reviewing the verification results, edit any names that showed up as
   unexpected (editing a name clears the verification badges and returns you to
   step 5). When everything looks right, click **Add to List**.
7. Aisle4 submits the selected items sequentially and shows a summary notice:
   - *"3 items added"* — newly added via POST
   - *"1 item updated"* — quantity appended to an existing list entry via PUT
   - *"1 failed"* — any item that got an unexpected error response (details
     logged to the console)

---

## File Layout

```
groceries/
  obsidian/aisle4/          ← plugin source (source of truth — in git)
    SPEC.md                 ← original build specification
    manifest.json           ← Obsidian plugin metadata
    styles.css              ← all plugin CSS
    main.js                 ← bundled output (generated — do not edit directly)
    src/
      main.js               ← Aisle4Plugin entry point
      settings.js           ← DEFAULT_SETTINGS + Aisle4SettingTab
      scraper.js            ← ingredient DOM scraping (pure functions)
      modal.js              ← Aisle4Modal — the review and verify dialog
      api.js                ← verifyItems() + addToGroceryList()

/Users/rnixon/Obsidian/merrett and ryan/.obsidian/plugins/aisle4/
                            ← deployed copy (not in git — do not edit here)
```

The vault copy is not in version control. Always edit in `groceries/obsidian/aisle4/src/`
and redeploy with `mise run plugin`.

---

## Build & Deploy

From the groceries repo root:

```
mise run plugin
```

This does two things:

1. Bundles `src/main.js` and its imports into `main.js` using **esbuild** with
   these flags:
   ```
   --bundle --platform=node --format=cjs --external:obsidian
   ```
   The `obsidian` module is marked external so it stays as `require('obsidian')`
   in the output — Obsidian injects it at runtime. CommonJS format is required
   because Obsidian's plugin loader uses `require()`.

2. Copies `main.js`, `manifest.json`, and `styles.css` to the vault plugin
   folder.

mise's `sources`/`outputs` tracking means the build step is skipped if no source
files changed since the last bundle. If you need to force a rebuild, delete
`main.js` before running `mise run plugin`.

After deploying, reload the plugin in Obsidian:

- **Settings → Community plugins** → disable and re-enable Aisle4, or
- **Cmd+P → "Reload app without saving"**

---

## Source Files

### `src/main.js` — Plugin entry point

Exports the `Aisle4Plugin` class (default export, required by Obsidian). Handles
`onload`, registers the ribbon icon and settings tab, and owns the top-level
`onRibbonClick` handler. This is where the scraper, modal, and API functions
are wired together:

```
ribbon click
  → findRecipeLeaf()        look for a Recipe View leaf for the active note
  → scrapeIngredients()     extract ingredients from the rendered DOM
  → Aisle4Modal(            open review dialog
      onVerify callback     → verifyItems()
      onSubmit callback     → addToGroceryList() + build summary Notice
    )
```

### `src/settings.js` — Settings tab

Defines `DEFAULT_SETTINGS` and `Aisle4SettingTab`. The settings tab renders the
API base URL, username, and password fields, plus a Connect button that calls
`POST /api/v1/auth/login` and stores the returned token. The connection status
(connected with expiry / expired / not connected) is displayed inline in the
same setting row.

Settings stored: `apiBaseUrl`, `username`, `password`, `token`, `tokenExpiresAt`.

### `src/scraper.js` — DOM scraper (pure functions)

Two public functions:

- **`findRecipeLeaf(app)`** — walks Obsidian's leaf tree looking for a leaf whose
  view type is `"recipe-view"` and whose file matches the currently active file.
  Returns the leaf, or `null` if not found.

- **`scrapeIngredients(contentEl)`** — detects whether the recipe-view DOM is
  in two-column or one-column layout, delegates to `extractIngredients()`, and
  returns an array of ingredient objects:
  ```js
  {
    group: string | null,   // sub-heading from the recipe (e.g. "Marinade")
    name: string,           // ingredient name, quantity spans removed
    quantity: string,       // primary quantity, possibly scaled
    original: string,       // full raw text before any processing
    doneInRecipeView: bool  // true if the user checked this off while cooking
  }
  ```

The `group` field is for modal display only — it must not be sent to the API.

### `src/modal.js` — Review modal

`Aisle4Modal` extends Obsidian's `Modal`. Constructor signature:

```js
new Aisle4Modal(app, ingredients, onVerify, onSubmit)
```

The modal maintains a mutable working copy of the ingredient list (so the
original scrape is never modified). Each item in the working copy carries
`included`, `status`, `listItemId`, `existingQuantity`, `statusEl`, and
`nameInputEl` fields alongside the data fields.

The modal has two modes:

- **`verify` mode** (initial) — the action button reads "Verify". Clicking it
  calls `onVerify`, shows status badges on each row, and transitions to submit
  mode. Editing any name input while in submit mode resets all badges and returns
  to verify mode.
- **`submit` mode** — the action button reads "Add to List". Clicking it filters
  to included, non-empty-name items, shows "Adding…" while awaiting `onSubmit`,
  then closes the modal.

Quantity edits do not reset the mode — quantities don't affect catalog matching.

### `src/api.js` — API helpers

Two exported functions:

**`verifyItems(items, settings)`**

Fetches the full item catalog (`GET /api/v1/items`) in a single call and
cross-references each ingredient name with a case-insensitive match. Returns an
array of result objects (one per item, in the same order):

```js
{
  status: 'known' | 'on-list' | 'new',
  canonicalName: string | null,     // API's stored casing, null for 'new'
  listItemId: number | null,        // catalog item ID, set for 'on-list' only
  existingQuantity: string | null,  // current list quantity, set for 'on-list' only
}
```

**`addToGroceryList(items, settings)`**

Iterates the selected items sequentially and submits each one:

- Items with `listItemId` set (i.e. `'on-list'` at verify time): `PUT
  /api/v1/list/items/{id}` with the existing and new quantities joined by `" + "`.
- All other items: `POST /api/v1/list/items` with `{ name, quantity }`. The
  server creates a new uncategorized catalog entry automatically if the name is
  not recognized.

Returns `{ added, appended, alreadyOnList, errors }`.

---

## API Endpoints Used

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `POST` | `/api/v1/auth/login` | None | Exchange credentials for a bearer token |
| `GET`  | `/api/v1/items`      | Bearer | Fetch full item catalog for verification |
| `POST` | `/api/v1/list/items` | Bearer | Add a new or unknown item to the shopping list |
| `PUT`  | `/api/v1/list/items/{id}` | Bearer | Update an existing list entry's quantity |

The `{id}` in the PUT path is the **catalog item ID** (from the `id` field of
the `GET /api/v1/items` response), not the list entry ID.

---

## Architecture & Design Decisions

### `requestUrl()` instead of `fetch()`

Obsidian runs in Electron. The renderer process enforces CORS just like a
browser, and the grocery API serves no `Access-Control-Allow-Origin` headers.
Native `fetch()` therefore silently fails. Obsidian's `requestUrl()` (from the
`obsidian` module) routes HTTP through the main Electron process where CORS is
not enforced.

All API calls in this plugin use `requestUrl()`. Key differences from `fetch()`:

- Pass `throw: false` to receive non-2xx responses as values rather than
  exceptions.
- Response body is accessed as `.json` (pre-parsed), `.text`, or
  `.arrayBuffer` — not streamed.

### DOM scraping instead of parsing raw markdown

Ingredients are scraped from the live rendered DOM produced by recipe-view, not
parsed from the raw markdown. The reason is quantity scaling: recipe-view lets
the user multiply servings (2×, 3×, etc.), and those scaled values only exist in
the DOM. Parsing the raw markdown would always capture the unscaled originals.

The coupling to recipe-view's internal DOM structure is a known trade-off.
recipe-view changes very rarely, and the DOM layout has been stable since the
plugin was written.

### Parenthetical alternate-unit quantities

Recipes often annotate ingredients with metric equivalents in parentheses —
`**1/4 cup (35 g)** garlic`. recipe-view wraps both quantities in `[data-qty]`
spans, leaving `(` and `)` as adjacent text nodes around the second span.

The `isParentheticalQty(span)` helper detects this by inspecting the span's
immediate text-node siblings. Parenthetical spans are excluded from the
`quantity` field and their surrounding parentheses are stripped from the `name`
field.

One subtlety: sibling references must be captured *before* any span is removed
from the DOM clone, because removing a node changes the sibling chain of
adjacent nodes. The scraper captures all `{ span, prev, next, parenthetical }`
tuples up front, then processes them.

### Two-step verify-then-submit flow

Submitting ingredients directly from scraped text is unreliable because:

1. The grocery API is case-sensitive. "Olive oil" and "olive oil" are different
   items to the server.
2. The user may want to know upfront which items are new vs already on the list
   before committing.
3. An ingredient already on the list should have its quantity *appended to*, not
   silently skipped — but blindly appending without user awareness would be
   surprising.

The verify step surfaces this information before any writes happen, gives the
user a chance to correct names, and gates submission behind a deliberate second
click.

### Single catalog fetch for verification

Rather than making one API call per ingredient to look up each name
individually, `verifyItems()` fetches the entire item catalog once and matches
client-side. The catalog is expected to be small (hundreds of items at most) so
the full payload is negligible, and a single round-trip is faster than a dozen.

### Case normalization at verify time

When the catalog produces a case-insensitive match, `handleVerify()` in the
modal silently overwrites the ingredient's name with the API's canonical casing
and updates the input field's displayed value. This happens automatically —
there is no notice or visual disruption — because it is purely a cosmetic
correction that the user would have had to make themselves anyway.

### Quantity appending for on-list items

When a verified `'on-list'` item is submitted, the plugin PUTs the catalog item
ID with a combined quantity string: `existingQuantity + " + " + newQuantity`.
The `appendQuantities()` helper handles blank quantities gracefully — if either
side is empty, the non-blank side is used directly without a stray `" + "`.

This is intentionally a string concatenation, not arithmetic. Quantities are
free-form text (e.g. `"2 cups"`, `"1/4 tsp"`) and there is no general way to
add them mathematically. A human reading `"2 cups + 1 cup"` understands what it
means and can clean it up on the grocery list if needed.

The `'on-list'` status badge is styled in red (rather than a softer warning
colour) to make it clear that submitting will modify an existing list entry, not
just add a new one.

### Sequential API calls

`addToGroceryList()` submits items one at a time in a `for` loop rather than
with `Promise.all`. With a typical recipe's ~12 ingredients the latency
difference is imperceptible, and sequential calls are much simpler to reason
about: each result is handled as it arrives and the final counters are
straightforward to accumulate.
