# iOS Shopping List by Store (Paged Top Tabs) - Design

Date: 2026-03-23
Status: Approved in brainstorming (ready for implementation planning)

## Overview

The iOS `Shopping List` screen currently renders all stores in a single long list with repeated store/category headings. As item volume grows, this becomes harder to scan and navigate.

This design restructures the screen into store-focused pages:

- A horizontal, scrollable store selector at the top (chip/pill style)
- One visible store per page, with horizontal swipe navigation between stores
- Store-level completion indicator on each selector chip when all items in that store are checked off

Only stores that currently have list items are shown in the selector/pages.

## Goals and Non-Goals

### Goals

- Reduce cognitive load by showing one store at a time
- Support many stores without cramped fixed tabs
- Preserve existing item interactions (toggle done, remove, add item, refresh)
- Keep bottom area available for future global navigation entries
- Show clear per-store completion state in the store selector

### Non-Goals

- No backend/API contract changes
- No changes to authentication/session flows
- No introduction of bottom tab navigation in this change
- No behavior change to global "Finish Shopping" semantics

## User Experience

### Store Navigation Model

- Top selector is horizontally scrollable to scale with many stores
- Tapping a store chip selects and navigates to that store page
- Swiping horizontally across pages updates selected chip
- Only non-empty stores are shown
- Store ordering for chips/pages follows `storeGroups` ordering from `ShoppingListViewModel` (currently alphabetical by store name); all "first available" selection behavior uses this deterministic order

### Store Completion Indicator

- A store chip shows a completion graphic (`checkmark.circle.fill`) only when:
  - the store has at least one item, and
  - all items in that store are `done == true`

### Selection Behavior

- Initial load with non-empty stores: select the first non-empty store ID
- Initial load with no non-empty stores: `selectedStoreID = nil` and show existing empty state
- On any store-ID set change (load/refresh/mutation):
  - If `selectedStoreID` still exists in non-empty store IDs, keep it
  - Else if non-empty store IDs is not empty, select the first available ID
  - Else set `selectedStoreID = nil`
- If current store becomes fully done but still has items, remain on that store (no auto-advance)
- If screen was empty and later receives non-empty stores, auto-select first available store ID

## Architecture and Components

## Existing Components Reused

- `ShoppingListViewModel` keeps owning load/mutation logic
- Existing `StoreGroup -> CategoryGroup -> ListItem` grouping model remains
- Existing row interactions and swipe actions remain
- Existing `AddItemBar`, error banner, toolbar, refresh, and loading/empty states remain

### Updated Components

#### `ShoppingListView`

- Replace "single running list of all stores" with:
  - top horizontal selector (`ScrollView(.horizontal)`) of store chips
  - paged store container implemented with `TabView(selection:)` + `.page(indexDisplayMode: .never)`
- Keep local selection state (`@State private var selectedStoreID: Int?`) as single UI source of truth
- Bind both chip tap selection and page selection to the same `selectedStoreID`
- Tag each store page with its `storeID` to keep chip/page synchronization deterministic
- Keep existing bottom safe-area inset for add-item + error banner

#### `ShoppingListViewModel`

Add derived helpers to support UI state and chip metadata:

- `nonEmptyStoreGroups: [StoreGroup]`
- `isStoreComplete(storeID: Int) -> Bool`
- `storeTotals(storeID: Int) -> (total: Int, done: Int)`

These are computed from current in-memory `items`/`storeGroups` without API changes.

## Data Flow

1. View loads via existing `load()` path (stores/categories/list/items in parallel)
2. View model rebuilds `storeGroups` as today
3. View reads `nonEmptyStoreGroups` to render selector + pages
4. `ShoppingListView` reconciles `selectedStoreID` whenever `nonEmptyStoreGroups.map(\.id)` changes:
   - keep selected ID when possible
   - otherwise fallback to first available
   - or set to nil when no stores remain
5. Item mutations (toggle/remove/add/finish) update `items` and rebuild groups
6. UI automatically reflects updated per-store completion/icon state

## Error Handling

- Keep existing inline error banner behavior
- No new modal or interruptive error handling
- Mutation errors continue to roll back optimistic state where applicable
- Selection fallback is deterministic and local; it does not produce user-facing errors

## Accessibility

- Store chips expose descriptive labels including completion context (e.g., "Target, 5 items, all done")
- Selected chip exposes selected trait/state
- Paging remains discoverable through VoiceOver swipe gestures
- Existing row accessibility labels/hints remain unchanged

## Testing Strategy

### ViewModel Tests

Add/extend tests for:

- Filtering to non-empty stores
- Per-store completion computation
- Per-store totals helper
- Deterministic ordering of non-empty stores (matching chosen ordering rule)

### View Selection-Reconciliation Tests

Add view/state tests for selection reconciliation logic in `ShoppingListView`:

- Preserve when selected store remains
- Fallback when selected store is removed
- Nil selection when all stores are empty
- Empty -> non-empty transition selects first available store
- No auto-switch when selected store is fully done but still non-empty

### UI/State Coverage

Add targeted previews and/or lightweight UI state tests for:

- Multiple stores with mixed completion
- Completed selected store (stays selected)
- Selected store removed and fallback applied
- Full empty-state transition

## Rollout Plan

1. Add view model helper properties/methods for non-empty stores and store completion metadata
2. Refactor `ShoppingListView` layout into top selector + paged store content
3. Add selection reconciliation to keep state stable across refresh/mutations
4. Wire chip completion icon logic and accessibility labels
5. Add/adjust tests and previews for critical state transitions

## Files Expected to Change

- `clients/ios/Sources/Groceries/Features/ShoppingList/ShoppingListView.swift`
- `clients/ios/Sources/Groceries/Features/ShoppingList/ShoppingListViewModel.swift`
- `clients/ios/Tests/GroceriesTests/Features/ShoppingList/ShoppingListViewModelTests.swift` (or existing equivalent feature-test target path)
- `clients/ios/Tests/GroceriesTests/Features/ShoppingList/ShoppingListViewStateTests.swift` (if view/state tests are used in this repo)

## Risks and Mitigations

- Risk: selection index/ID drift when store set changes quickly
  - Mitigation: use store ID as source of truth; reconcile on store-ID set changes
- Risk: paging container and chip selection can desync
  - Mitigation: bind both to the same `selectedStoreID` state
- Risk: regressions in existing row mutation UX
  - Mitigation: preserve row component and mutation paths, and verify behavior with tests

## Success Criteria

- Users can switch between stores via top selector and horizontal paging
- Only stores with items appear in selector/pages
- Completion icon appears only for stores with all items done
- Current store does not auto-advance just because all items are done
- Existing list interactions and global finish-shopping flow continue to work
