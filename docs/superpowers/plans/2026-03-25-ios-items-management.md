# iOS Items Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a dedicated iOS Items tab (between List and Account) with add/edit/delete and edit-screen-only shopping-list membership toggling, including deterministic Shopping List consistency after membership changes.

**Architecture:** Introduce an isolated `Features/Items` module with its own view model and editor/add views, keep `ShoppingListViewModel` independent, and synchronize membership changes via a shared `NotificationCenter` event plus a deduped refresh coordinator path in `ShoppingListView`. Add explicit GroceriesAPI item endpoints in a dedicated `ItemEndpoints.swift` file.

**Tech Stack:** Swift 6, SwiftUI, Observation (`@Observable`), async/await, XCTest, Tuist, URLSession + `MockURLProtocol`.

---

## File Map

### Create

- `clients/ios/Sources/GroceriesAPI/ItemEndpoints.swift` — item list/create/update/delete endpoint wrappers.
- `clients/ios/Sources/Groceries/Shared/AppEvents.swift` — notification names + payload keys for cross-feature sync.
- `clients/ios/Sources/Groceries/Features/Items/ItemsViewModel.swift` — item screen state, filtering, mutations.
- `clients/ios/Sources/Groceries/Features/Items/ItemsView.swift` — item list/search/filter and navigation entry points.
- `clients/ios/Sources/Groceries/Features/Items/AddItemView.swift` — required category + name item creation form.
- `clients/ios/Sources/Groceries/Features/Items/ItemEditorView.swift` — edit name/category, toggle list membership, delete confirmation.
- `clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift` — feature logic tests.

### Modify

- `clients/ios/Sources/GroceriesAPI/Models.swift` — add `CreateItemRequest` and `UpdateItemRequest`.
- `clients/ios/Sources/Groceries/Features/Navigation/AppTabsView.swift` — insert `items` tab between `list` and `account`.
- `clients/ios/Sources/Groceries/Features/ShoppingList/ShoppingListView.swift` — subscribe to membership event and refresh through deduped path.
- `clients/ios/Tests/GroceriesAPITests/GroceriesAPIClientTests.swift` — endpoint/query/error-mapping coverage.
- `clients/ios/Tests/GroceriesTests/Features/Navigation/AppNavigationTests.swift` — tab-order assertion update.
- `clients/ios/Tests/GroceriesTests/Features/ShoppingList/ShoppingListAutoRefreshCoordinatorTests.swift` — dedupe/race refresh coverage.
- `clients/ios/README.md` — update architecture + future-work sections for new Items feature.

---

### Task 1: Add API request models (TDD)

**Files:**
- Modify: `clients/ios/Tests/GroceriesAPITests/GroceriesAPIClientTests.swift`
- Modify: `clients/ios/Sources/GroceriesAPI/Models.swift`

- [ ] **Step 1: Write failing model-encoding tests**

```swift
func testEncodeCreateItemRequest() throws {
    let body = CreateItemRequest(categoryID: 3, name: "Apples")
    let data = try JSONEncoder.apiEncoder.encode(body)
    let json = try XCTUnwrap(String(data: data, encoding: .utf8))
    XCTAssertTrue(json.contains("\"category_id\":3"))
    XCTAssertTrue(json.contains("\"name\":\"Apples\""))
}

func testEncodeUpdateItemRequest() throws {
    let body = UpdateItemRequest(categoryID: 9, name: "Oat Milk")
    let data = try JSONEncoder.apiEncoder.encode(body)
    let json = try XCTUnwrap(String(data: data, encoding: .utf8))
    XCTAssertTrue(json.contains("\"category_id\":9"))
    XCTAssertTrue(json.contains("\"name\":\"Oat Milk\""))
}
```

- [ ] **Step 2: Run tests to verify RED**

Run: `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesAPI -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesAPITests/GroceriesAPIClientTests/testEncodeCreateItemRequest -only-testing:GroceriesAPITests/GroceriesAPIClientTests/testEncodeUpdateItemRequest`
Expected: FAIL due to missing request model types.

- [ ] **Step 3: Implement minimal request models**

```swift
public struct CreateItemRequest: Encodable, Sendable {
    public let categoryID: Int
    public let name: String

    enum CodingKeys: String, CodingKey {
        case categoryID = "category_id"
        case name
    }
}

public struct UpdateItemRequest: Encodable, Sendable {
    public let categoryID: Int
    public let name: String

    enum CodingKeys: String, CodingKey {
        case categoryID = "category_id"
        case name
    }
}
```

- [ ] **Step 4: Run tests to verify GREEN**

Run the same command as Step 2.
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add clients/ios/Sources/GroceriesAPI/Models.swift clients/ios/Tests/GroceriesAPITests/GroceriesAPIClientTests.swift
git commit -m "test: add item request model encoding coverage"
```

---

### Task 2: Add item API endpoints and API tests (TDD)

**Files:**
- Create: `clients/ios/Sources/GroceriesAPI/ItemEndpoints.swift`
- Modify: `clients/ios/Tests/GroceriesAPITests/GroceriesAPIClientTests.swift`

- [ ] **Step 1: Write failing endpoint tests**

Add tests for:
- `testListItems_withInListQueryEncodesQueryItem`
- `testCreateItem_postsExpectedBody`
- `testUpdateItem_putsExpectedBody`
- `testDeleteItem_sendsDelete`
- `testDeleteItem_conflict_throws409`
- `testUpdateItem_notFound_throws404`
- `testDeleteItem_notFound_throws404`

Use `MockURLProtocol.setRequestHandler` to assert method/path/query/body.

- [ ] **Step 2: Run tests to verify RED**

Run: `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesAPI -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesAPITests/GroceriesAPIClientTests`
Expected: FAIL for missing methods in `GroceriesAPIClient` extension.

- [ ] **Step 3: Implement minimal endpoint wrappers**

```swift
public func listItems(categoryID: Int? = nil, inList: Bool? = nil) async throws -> [Item]
public func createItem(categoryID: Int, name: String) async throws -> Item
public func updateItem(id: Int, categoryID: Int, name: String) async throws -> Item
public func deleteItem(id: Int) async throws
```

Build query items only when parameters are non-nil.

- [ ] **Step 4: Run tests to verify GREEN**

Run the same command as Step 2.
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add clients/ios/Sources/GroceriesAPI/ItemEndpoints.swift clients/ios/Tests/GroceriesAPITests/GroceriesAPIClientTests.swift
git commit -m "feat: add iOS item management API endpoints"
```

---

### Task 3: Add shared app event contract and navigation tab order (TDD)

**Files:**
- Create: `clients/ios/Sources/Groceries/Shared/AppEvents.swift`
- Modify: `clients/ios/Sources/Groceries/Features/Navigation/AppTabsView.swift`
- Modify: `clients/ios/Tests/GroceriesTests/Features/Navigation/AppNavigationTests.swift`

- [ ] **Step 1: Write failing navigation and event-contract tests**

Add/adjust tests:
- `test_appTabsIncludeListItemsAndAccount`
- `test_itemsMembershipDidChangeNotificationPayloadKeys`

```swift
XCTAssertEqual(AppTab.allCases, [.list, .items, .account])
XCTAssertEqual(AppEvents.MembershipChanged.itemIDKey, "itemID")
```

- [ ] **Step 2: Run tests to verify RED**

Run: `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesTests/AppNavigationTests`
Expected: FAIL due to missing `items` tab/event constants.

- [ ] **Step 3: Implement tab enum update and shared event constants**

Create constants similar to:

```swift
enum AppEvents {
    enum MembershipChanged {
        static let name = Notification.Name("itemsMembershipDidChange")
        static let itemIDKey = "itemID"
        static let isInListKey = "isInList"
        static let changedAtKey = "changedAt"
    }
}
```

Insert the Items tab between existing List and Account tabs.

- [ ] **Step 4: Run tests to verify GREEN**

Run command from Step 2.
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add clients/ios/Sources/Groceries/Shared/AppEvents.swift clients/ios/Sources/Groceries/Features/Navigation/AppTabsView.swift clients/ios/Tests/GroceriesTests/Features/Navigation/AppNavigationTests.swift
git commit -m "feat: add items tab and shared app event contract"
```

---

### Task 4: Implement ItemsViewModel with filtering and mutation logic (TDD)

**Files:**
- Create: `clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift`
- Create: `clients/ios/Sources/Groceries/Features/Items/ItemsViewModel.swift`

- [ ] **Step 1: Write failing ViewModel tests for search/filter and validation**

Add tests for:
- case-insensitive substring filtering
- in-list-only + search composition
- trimmed-empty name validation on add/edit
- required category validation
- initial load failure keeps empty-safe state and exposes retry path

- [ ] **Step 2: Run tests to verify RED**

Run: `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesTests/ItemsViewModelTests`
Expected: FAIL because `ItemsViewModel` and APIs are missing.

- [ ] **Step 3: Implement minimal ItemsViewModel state + pure filtering**

Implement:
- `items`, `filteredItems`, `searchText`, `inListOnly`, `categories`, `errorMessage`, and in-flight flags.
- `applyFilters()` and validation helpers.

- [ ] **Step 4: Add failing tests for mutation semantics and event emission**

Add tests for:
- add/edit/delete success mutating local cache
- toggle membership success emits notification with correct payload types
- toggle/delete failure preserves visible state
- failed toggle does not emit membership-change notification
- duplicate-submit protection while in flight
- manual retry re-fetches categories + items and clears stale cache
- pull-to-refresh re-fetches categories + items and replaces stale cache

- [ ] **Step 5: Run tests to verify RED**

Run command from Step 2.
Expected: FAIL for missing mutation implementations.

- [ ] **Step 6: Implement minimal mutation behavior**

Implement `addItem`, `updateItem`, `deleteItem`, `setInList` with:
- pessimistic commit
- in-flight gates
- notification posting on membership success only

- [ ] **Step 7: Run tests to verify GREEN**

Run command from Step 2.
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add clients/ios/Sources/Groceries/Features/Items/ItemsViewModel.swift clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift
git commit -m "feat: add items view model with filtering and mutations"
```

---

### Task 5: Build Items list and Add view UI (TDD for behavior-driving helpers)

**Files:**
- Create: `clients/ios/Sources/Groceries/Features/Items/ItemsView.swift`
- Create: `clients/ios/Sources/Groceries/Features/Items/AddItemView.swift`
- Modify: `clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift`

- [ ] **Step 1: Add failing tests for Add flow behavior exposed by ViewModel**

Test behaviors:
- add button disabled while add request in-flight
- category picker disabled while add request in-flight
- name field disabled while add request in-flight
- add requires non-empty trimmed name and category

- [ ] **Step 2: Run tests to verify RED**

Run: `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesTests/ItemsViewModelTests`
Expected: FAIL for missing `isAdding`/validation gates.

- [ ] **Step 3: Implement minimal behavior in ViewModel and wire UI**

Build `ItemsView` with:
- search field
- `In List only` toggle
- list of filtered items
- add button opening `AddItemView`

Build `AddItemView` with:
- category picker (required)
- name field (required)
- save/cancel actions
- category picker + name field disabled while add request is in-flight
- required accessibility labels for key controls

- [ ] **Step 4: Run targeted tests + compile checks**

Run:
- `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesTests/ItemsViewModelTests`
- `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17'`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add clients/ios/Sources/Groceries/Features/Items/ItemsView.swift clients/ios/Sources/Groceries/Features/Items/AddItemView.swift clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift
git commit -m "feat: add items list screen and add-item flow"
```

---

### Task 6: Build ItemEditor view with toggle/delete semantics (TDD)

**Files:**
- Create: `clients/ios/Sources/Groceries/Features/Items/ItemEditorView.swift`
- Modify: `clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift`

- [ ] **Step 1: Write failing tests for editor mutation matrix and failure semantics**

Add tests for:
- save/toggle/delete lock each other while in flight
- delete 409 keeps editor state and shows error
- 404 failure keeps prior state

- [ ] **Step 2: Run tests to verify RED**

Run: `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesTests/ItemsViewModelTests`
Expected: FAIL because state rules are not fully implemented.

- [ ] **Step 3: Implement ItemEditorView + ViewModel gates**

Implement editor UI with:
- name and category editable fields
- in-list toggle (editor only)
- delete confirmation dialog
- disable controls based on active mutation

- [ ] **Step 4: Run tests to verify GREEN**

Run command from Step 2.
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add clients/ios/Sources/Groceries/Features/Items/ItemEditorView.swift clients/ios/Tests/GroceriesTests/Features/Items/ItemsViewModelTests.swift
git commit -m "feat: add item editor with toggle and delete confirmation"
```

---

### Task 7: Wire Shopping List refresh coordination and dedupe (TDD)

**Files:**
- Modify: `clients/ios/Sources/Groceries/Features/ShoppingList/ShoppingListView.swift`
- Modify: `clients/ios/Tests/GroceriesTests/Features/ShoppingList/ShoppingListAutoRefreshCoordinatorTests.swift`

- [ ] **Step 1: Write failing tests for deduped refresh behavior**

Add tests for:
- notification burst collapses to one refresh request
- notification during in-flight refresh does not trigger overlapping refresh
- on-appear trigger and notification trigger share same gate logic
- notification + immediate tab switch race results in exactly one refresh
- repeated appear/disappear cycles do not register duplicate notification observers

- [ ] **Step 2: Run tests to verify RED**

Run: `xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17' -only-testing:GroceriesTests/ShoppingListAutoRefreshCoordinatorTests`
Expected: FAIL due to missing trigger-coordination behavior.

- [ ] **Step 3: Implement minimal refresh coordinator integration**

Update `ShoppingListView` to route both:
- `onAppear` refresh request
- `AppEvents.MembershipChanged.name` notification refresh request

through one deduped refresh entry point that respects `isRefreshingList` and 300ms dedupe.

- [ ] **Step 4: Run tests to verify GREEN**

Run command from Step 2.
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add clients/ios/Sources/Groceries/Features/ShoppingList/ShoppingListView.swift clients/ios/Tests/GroceriesTests/Features/ShoppingList/ShoppingListAutoRefreshCoordinatorTests.swift
git commit -m "fix: refresh shopping list after item membership changes"
```

---

### Task 8: Integrate app navigation and run full iOS test suite

**Files:**
- Modify: `clients/ios/README.md`

- [ ] **Step 1: Verify Items tab integration from Task 3 is complete**

Ensure `TabView` order is:
1) List
2) Items
3) Account

and `ItemsView(apiClient:)` receives the authenticated client.

No additional functional edits to `AppTabsView.swift` should be needed in this task; this step is verification-only to avoid churn.

- [ ] **Step 2: Run full API and app tests**

Run:

```bash
xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesAPI -destination 'platform=iOS Simulator,name=iPhone 17'
xcodebuild test -project clients/ios/Groceries.xcodeproj -scheme GroceriesTests -destination 'platform=iOS Simulator,name=iPhone 17'
```

Expected: PASS for both schemes.

- [ ] **Step 3: Run formatting/lint checks used by project workflow (if configured)**

Run: `mise run lint` (repo root)
Expected: PASS (or document any unrelated pre-existing failures).

- [ ] **Step 4: Update iOS README feature notes**

Document new Items feature in architecture/future-work sections so docs match shipped behavior.

- [ ] **Step 5: Run manual simulator verification checklist**

In simulator, verify:
1) Add item requires category and name.
2) Edit name/category saves correctly.
3) Toggle in-list in editor reflects in Shopping List after returning.
4) Delete requires confirmation and removes item.
5) Force failure paths (e.g., API unavailable/4xx) preserve previous visible state and show error.
6) Failed membership toggle does not trigger cross-tab refresh side effects.
7) Add and Edit controls include clear accessibility labels and disable correctly during in-flight mutations.
8) Delete action is explicitly destructive and requires confirmation before execution.

- [ ] **Step 6: Commit**

```bash
git add clients/ios/README.md
git commit -m "feat: integrate items management tab into iOS app"
```

---

## Final Verification Checklist

- [ ] `GroceriesAPI` tests pass locally.
- [ ] `GroceriesTests` pass locally.
- [ ] Tab order is exactly List, Items, Account.
- [ ] Add/Edit require category and non-empty name.
- [ ] In-list toggle exists only in editor and is pessimistic.
- [ ] Delete uses confirmation dialog and handles conflict errors.
- [ ] Membership toggle updates Shopping List view via event + deduped refresh path.
- [ ] README reflects new feature structure.

---

## Notes for Execution

- Follow @superpowers/test-driven-development on each task.
- Keep commits small and task-scoped.
- Prefer minimal implementation per test; avoid speculative abstractions.
- If API contracts differ from OpenAPI during implementation, update tests first, then adjust endpoint code.
