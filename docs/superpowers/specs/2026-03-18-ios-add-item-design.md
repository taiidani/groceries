# iOS Add Item to Shopping List — Design Spec

**Date:** 2026-03-18
**Status:** Approved

---

## 1. Scope & Goals

This feature adds the ability to add items to the shopping list from within the iOS app. It is the first significant UI capability built on top of the existing SwiftUI skeleton.

### In scope

- Display the current shopping list (already implemented; this feature extends it)
- Add an existing catalogue item to the list by searching and selecting it
- Create a new uncategorized item by free-text name and immediately add it to the list
- Capture an optional quantity string as part of every add operation

### Out of scope

- Editing the item catalogue (categories, stores)
- Editing quantity after an item has been added
- Recipe support
- Real-time SSE updates
- A mock protocol boundary for `ShoppingListViewModel` unit tests

---

## 2. Architecture & File Structure

The existing SwiftUI layer (`Sources/Groceries/`) and API framework (`Sources/GroceriesAPI/`) are both present. No new targets or dependencies are required.

### New file

```
Sources/Groceries/Features/ShoppingList/AddItemBar.swift
```

A self-contained SwiftUI view that owns all transient UI state for the add flow: search text, search results, selected item, quantity, and in-flight loading state. It surfaces a single async callback to the view model:

```swift
onAdd: (_ itemID: Int?, _ name: String?, _ quantity: String) async throws -> Void
```

### Modified files

| File | Change |
|---|---|
| `Sources/GroceriesAPI/ListEndpoints.swift` | Add `listItems()` wrapping `GET /api/v1/items` |
| `Features/ShoppingList/ShoppingListViewModel.swift` | Add `searchItems(query:)` and `addItem(itemID:name:quantity:)` |
| `Features/ShoppingList/ShoppingListView.swift` | Wire `AddItemBar` into `.safeAreaInset(edge: .bottom)` |

### Unchanged files

`GroceriesApp.swift`, `Auth/`, `GroceriesAPI/Models.swift`, `GroceriesAPI/StoreEndpoints.swift`

---

## 3. Interaction Flow

### 3.1 Idle state

`ShoppingListView` renders `AddItemBar` via `.safeAreaInset(edge: .bottom)`. The bar is always visible as a tappable "Add an item…" pill. It sits above the home indicator and, in future, above the tab bar — `.safeAreaInset` stacks correctly in both cases with no additional work.

### 3.2 Active state — search

Tapping the bar focuses a text field and the keyboard rises. SwiftUI's built-in keyboard avoidance scrolls the list content up automatically.

As the user types, `searchItems(query:)` filters the locally cached item list by case-insensitive name prefix. No API call is made per keystroke — the full item list is fetched once during `load()` and cached in the view model. Results are displayed as a scrollable list of rows showing item name and category name.

When the search field is non-empty, a **"Add `<typed text>` as new item"** row is appended below the results. When the field is empty, no API call is made and no results are shown.

A **Cancel** button dismisses the keyboard and resets the bar to idle.

### 3.3 Selection — quantity step

Tapping a result row (existing item or new-item row) selects it. The results list collapses and a quantity input field expands inline below the selected item's name chip. Keyboard focus moves to the quantity field. An **Add** button appears to the right of the quantity field.

### 3.4 Confirm

Tapping **Add** (or Return on the quantity field) invokes `addItem(itemID:name:quantity:)` on the view model. On success the bar resets to idle and the new `ListItem` is appended to `items` and surfaced via `rebuildGroups()` — no full list reload is needed.

### 3.5 Cancel from quantity step

Tapping **Cancel** at the quantity step dismisses the keyboard and resets the bar to idle without adding anything.

---

## 4. View Model Changes (`ShoppingListViewModel`)

### New stored property: `allItems: [Item]`

`private(set) var allItems: [Item] = []`

Populated as a fourth parallel task in `load()` alongside stores, categories, and the shopping list. Not re-fetched on pull-to-refresh.

### `searchItems(query: String) -> [Item]`

- Synchronously filters `allItems` by case-insensitive name prefix match
- Returns an empty array when `query` is empty
- No async, no API call

### `addItem(itemID: Int?, name: String?, quantity: String) async throws`

- Dispatches to `apiClient.addItemToList(itemID:quantity:)` when `itemID` is non-nil
- Dispatches to `apiClient.addNewItemToList(name:quantity:)` when `name` is non-nil
- On success: appends the returned `ListItem` to `items`, increments `total`, calls `rebuildGroups()`
- On failure: sets `errorMessage` via the existing error description helper
- Uses `mutatingItemIDs` is not appropriate here (no existing item to track); instead `AddItemBar` owns its own `isAdding: Bool` state to disable the Add button during the request

---

## 5. API Layer Changes (`GroceriesAPI`)

### `ListEndpoints.swift` — new method

```swift
public func listItems() async throws -> [Item]
```

Calls `GET /api/v1/items` with no query parameters, returning all items. The server already loads all items and filters client-side in its own handler — adding a name filter query parameter would require a backend change and is unnecessary given the item catalogue is expected to be small (hundreds, not thousands). The iOS app caches this result in `ShoppingListViewModel.allItems` and filters locally.

The full item list is fetched once as part of the parallel `load()` call alongside stores, categories, and the shopping list. It is not refreshed on pull-to-refresh (items change rarely).

> **Note:** The `Item` model already exists in `Models.swift`.

---

## 6. UI Placement & HIG Compliance

| Concern | Decision |
|---|---|
| Entry point | `.safeAreaInset(edge: .bottom)` — always visible, keyboard-aware, tab-bar-safe |
| Toolbar conflicts | None — existing top-left spinner and top-right "Finish Shopping" button are untouched |
| Navigation | No push or sheet — fully inline within `ShoppingListView` |
| Keyboard avoidance | Handled automatically by SwiftUI |
| Dark mode | `AddItemBar` inherits `.environment(\.colorScheme, .dark)` from the existing list view |
| Accessibility | Search field and Add button require `accessibilityLabel`; results rows require combined labels including category name |

---

## 7. Error Handling

| Scenario | Behaviour |
|---|---|
| `listItems()` fetch failure during `load()` | Sets `errorMessage` via the existing error banner; search field will show no results until a successful reload |
| Add failure — generic | Sets `errorMessage`; surfaces via existing error banner |
| Add failure — 409 Conflict (already on list) | `APIError.conflict` message displayed in error banner |
| Add failure — 404 Not Found | `APIError.notFound` message displayed in error banner |

---

## 8. Testing

New tests added to `Tests/GroceriesAPITests/`:

| Test | Coverage |
|---|---|
| `testListItems_returnsAllItems` | Verifies `listItems()` calls `GET /api/v1/items` and decodes the response |
| `testAddItemToList_existingItem` | Verifies correct JSON body and `ListItem` decoding |
| `testAddNewItemToList_freeText` | Verifies free-text path sends correct JSON body |
| `testAddToList_conflict_throws409` | Verifies 409 maps to `APIError.conflict` |

`searchItems(query:)` is a pure synchronous filter and can be tested directly on `ShoppingListViewModel` without a network mock:

| Test | Coverage |
|---|---|
| `testSearchItems_filtersByPrefix` | Verifies case-insensitive prefix match returns correct subset of `allItems` |
| `testSearchItems_emptyQueryReturnsEmpty` | Verifies empty query returns empty array |

`ShoppingListViewModel` is not unit-tested in this feature (requires a mock protocol boundary — deferred to future work). Behaviour is covered by API-layer tests and simulator testing.
