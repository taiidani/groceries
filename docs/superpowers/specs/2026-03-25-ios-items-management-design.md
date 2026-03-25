# iOS Items Management — Design Spec

**Date:** 2026-03-25
**Status:** Approved

---

## 1. Scope & Goals

This feature adds a dedicated Items section to the iOS client for item-catalog management, matching the existing webapp capability pattern while preserving iOS-specific flow and usability.

### In scope

- Add a new item with required category selection.
- View and search/filter a long list of items.
- Open a dedicated edit screen per item.
- Update item name and category.
- Toggle whether an item is in the shopping list (edit screen only).
- Delete an item with destructive confirmation.
- Place the Items tab between List and Account in bottom navigation.
- Ensure Shopping List reflects in/out-of-list changes made from Items.

### Out of scope

- Bulk-edit operations from the item list screen.
- Editing item quantity from the Items feature.
- Replacing existing Shopping List architecture with a fully shared global store.
- SSE-based live synchronization for the Items feature.

---

## 2. Architecture & Navigation

The feature is implemented as a dedicated module under `Features/Items` with its own view model and navigation stack.

- Add a new `items` tab in `AppTabsView` between `list` and `account`.
- `ItemsView` owns browse/search/filter experience.
- Tapping an item pushes `ItemEditorView` so edit controls have full-screen space.
- Add action opens `AddItemView` to create a new catalog item.
- `ShoppingListView` remains independent and refreshes when item list-membership changes are published.

This keeps item-management concerns isolated from shopping-list concerns while still providing deterministic cross-tab consistency.

---

## 3. File Structure

### New files

- `clients/ios/Sources/Groceries/Features/Items/ItemsView.swift`
- `clients/ios/Sources/Groceries/Features/Items/ItemsViewModel.swift`
- `clients/ios/Sources/Groceries/Features/Items/AddItemView.swift`
- `clients/ios/Sources/Groceries/Features/Items/ItemEditorView.swift`

### Modified files

- `clients/ios/Sources/Groceries/Features/Navigation/AppTabsView.swift`
  - Add `AppTab.items` and place its tab item between List and Account.
- `clients/ios/Sources/Groceries/Features/ShoppingList/ShoppingListView.swift`
  - Listen for item-membership-change event and refresh.
  - Refresh on re-appear safety path.
- `clients/ios/Sources/Groceries/Shared/AppEvents.swift` (new shared constants file)
  - Define `Notification.Name` constants and payload keys used for cross-feature sync.
- `clients/ios/Sources/GroceriesAPI/Models.swift`
  - Add `CreateItemRequest` and `UpdateItemRequest` request models.
- `clients/ios/Sources/GroceriesAPI/ItemEndpoints.swift` (new)
  - Add item-management endpoint wrappers for list/create/update/delete.

### Test files added/updated

- `clients/ios/Tests/GroceriesAPITests/GroceriesAPIClientTests.swift`
- `clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift` (new)
- `clients/ios/Tests/GroceriesTests/Features/Navigation/AppNavigationTests.swift`

---

## 4. Data & Endpoint Contract

The iOS client uses existing backend item and list endpoints from `openapi.yaml`.

### Item catalog endpoints

- `GET /api/v1/items`
- `POST /api/v1/items`
- `PUT /api/v1/items/{id}`
- `DELETE /api/v1/items/{id}`

### Concrete iOS client API contract

`ItemEndpoints.swift` will define the exact public methods below:

```swift
public func listItems(categoryID: Int? = nil, inList: Bool? = nil) async throws -> [Item]
public func createItem(categoryID: Int, name: String) async throws -> Item
public func updateItem(id: Int, categoryID: Int, name: String) async throws -> Item
public func deleteItem(id: Int) async throws
```

### Filtering strategy (v1)

- V1 uses **client-side filtering only** for predictable UX and low latency while typing.
- `ItemsViewModel` fetches the full item catalog once on screen load and filters in memory.
- Text search and `In List only` toggle are composed in memory.
- Query-param-based server filtering (`category_id`, `in_list`) remains available in API but is not used in v1 UI flow.

### Cache freshness policy

- Initial load: fetch categories + items in parallel.
- Manual retry path: re-fetch categories + items.
- Pull-to-refresh in `ItemsView`: re-fetch categories + items.
- App foreground: no automatic fetch in v1 (avoid surprise network churn).
- Mutation success updates local cache immediately (add/edit/delete/toggle).
- Mutation failure does not mutate cache; error is shown and user may retry.

### Shopping-list membership endpoints

- Add to list: `POST /api/v1/list/items` with `item_id`.
- Remove from list: `DELETE /api/v1/list/items/{itemID}`.

### iOS request models

```swift
public struct CreateItemRequest: Encodable, Sendable {
    public let categoryID: Int
    public let name: String
}

public struct UpdateItemRequest: Encodable, Sendable {
    public let categoryID: Int
    public let name: String
}
```

---

## 5. Interaction Flow

### 5.1 Items list screen

- On first appear, load categories and items.
- Show search field and `In List only` toggle.
- Default list mode: all items.
- Search is client-side, case-insensitive substring against item name.
- `In List only` filter composes with text search.
- Filtering is applied in-memory against cached full catalog data (no per-keystroke API calls).

### 5.2 Add item flow

- User taps add action from Items screen.
- `AddItemView` requires category + name.
- Save calls create endpoint.
- On success, dismiss and update local item cache.

### 5.3 Edit item flow

- User selects an item from list.
- `ItemEditorView` shows:
  - editable name
  - editable category
  - `In Shopping List` toggle (only place this toggle exists)
  - destructive delete button
- Save for name/category uses item update endpoint.
- Toggle ON uses add-to-list endpoint.
- Toggle OFF uses remove-from-list endpoint.

### 5.4 Delete flow

- Delete action presents confirmation dialog.
- Confirm executes item delete endpoint.
- On success, pop back to list and remove item from cache.

---

## 6. Cross-Tab Consistency Contract

When list-membership changes in `ItemEditorView`, the Shopping List tab must reflect that change.

### Chosen strategy

- Use event-based synchronization (`NotificationCenter`) with a shared constant:
  - Name: `Notification.Name.itemsMembershipDidChange`
  - Defined in `Shared/AppEvents.swift`
- Notification payload (`userInfo`) contract:
  - `itemID: Int`
  - `isInList: Bool`
  - `changedAt: Date`
- Emission rule:
  - Emit **only after a successful** membership mutation response.
  - Emit on main actor after local editor state is updated.
- Observer behavior:
  - `ShoppingListView` subscribes on appear and unsubscribes on disappear.
  - On event receipt, request refresh through a small refresh coordinator.
- Coalescing rule:
  - If multiple triggers arrive within 300ms, collapse to one refresh request.
- Navigation safety net:
  - `ShoppingListView` triggers `refreshIfNeeded()` on appear so returning from Items always revalidates list state.
  - Tab reselection (tapping the already-active List tab) is not required in v1 because standard SwiftUI `TabView` does not provide a stable built-in reselection signal without UIKit bridging.

### Refresh coordinator rules

- `ShoppingListView` owns a lightweight coordinator with:
  - `isRefreshing` gate (ignore additional triggers while refresh is in flight).
  - `lastRefreshRequestAt` timestamp for 300ms dedupe window.
- Trigger sources:
  - membership-change notification
  - `onAppear` safety refresh
- Both trigger sources call the same coordinator entry point to guarantee deterministic behavior in notification + navigation races.

This provides deterministic visible consistency without introducing a shared global store in this iteration.

---

## 7. Error Handling & UX Rules

- Keep pessimistic commit behavior for mutating actions.
  - Show in-flight state while request is pending.
  - Update UI state only after request success.
- Reuse existing API error mapping and present clear inline feedback.
- Validation:
  - Name is required and trimmed.
  - Category selection is required for add and edit.
- Delete conflict (`409`) keeps editor open and displays the returned message.
- Disable Save/Delete/toggle controls while their corresponding request is in flight.
- Accessibility labels and destructive semantics are required for key controls.

### Mutation concurrency matrix

- While **save name/category** is in flight:
  - Disable save button, fields, list-membership toggle, and delete button.
- While **toggle membership** is in flight:
  - Disable toggle control, save button, fields, and delete button.
- While **delete** is in flight:
  - Disable all controls and keep confirmation dialog dismissed.
- While **add item** is in flight (`AddItemView`):
  - Disable save button and form controls (name/category).
- Repeated taps on a disabled control are ignored.
- No cancellation path is exposed for in-flight item mutations in v1.

### Failure semantics

- `404` on save/toggle/delete:
  - Show server message.
  - Keep editor visible.
  - Offer retry; no optimistic local cache mutation is applied.
- `409` on delete (in-use item):
  - Show conflict message from server.
  - Keep editor visible.
  - No local cache mutation.
- Generic network/server failures:
  - Show inline error.
  - Preserve prior visible state (pessimistic commit).

---

## 8. Testing Strategy

### API-level tests

- Verify `createItem` request method/path/body and decoding.
- Verify `updateItem` request method/path/body and decoding.
- Verify `deleteItem` request method/path.
- Verify `listItems(categoryID:inList:)` query encoding and decoding.
- Verify `404/409` status mapping for item mutation endpoints.

### ViewModel tests

- Search filtering behavior (case-insensitive substring).
- `In List only` filter composition with search.
- Add success updates local items.
- Add flow prevents duplicate submit while request is in flight.
- Edit success updates local fields.
- Membership toggle updates state only on success.
- Membership toggle failure preserves prior UI state and surfaces error.
- Delete success removes item.
- Delete conflict surfaces error and preserves editor/list state.
- Add/Edit validation rejects trimmed-empty names.
- Add/Edit validation requires category selection.
- Rapid repeated toggle attempts do not send overlapping mutations.
- Initial load failure keeps empty-safe UI and allows manual retry.
- Pull-to-refresh replaces stale local cache with latest categories/items.

### Navigation tests

- Ensure tab order is `List`, `Items`, `Account`.
- Ensure returning to Shopping List view from Items triggers the on-appear refresh path.

### Cross-feature sync tests

- Verify membership-change notification posts only on successful toggle.
- Verify failed toggle does not emit notification.
- Verify notification payload contract includes `itemID`, `isInList`, and `changedAt` with expected value types.
- Verify Shopping List reacts to notification by refreshing once when notifications are bursty.
- Verify notification + immediate tab switch race still results in exactly one refresh.
- Verify notification received during in-flight refresh is deduped (no overlapping refreshes).

### Simulator verification

- Add item (category required).
- Edit name and category.
- Toggle in-list from editor and verify Shopping List reflects change.
- Delete with confirmation dialog.
- Confirm failed toggle/save/delete preserves previous visible state with clear error message.

---

## 9. Non-Goals and Follow-Up Options

Possible future iterations, explicitly deferred from this feature:

- Shared observable app store for richer cross-feature sync.
- Bulk item management actions.
- Server-driven filtered item pagination for very large catalogs.
- SSE integration for live item/list consistency.
