import Foundation
import GroceriesAPI

// MARK: - Hierarchy types

/// A category with its list items, used for display grouping.
public struct CategoryGroup: Identifiable {
    public let id: Int
    public let name: String
    public let items: [ListItem]
}

/// A store with its populated category groups, used for display grouping.
public struct StoreGroup: Identifiable {
    public let id: Int
    public let name: String
    public let categories: [CategoryGroup]
}

// MARK: - ShoppingListViewModel

/// Manages the shopping list data and coordinates all list-related API calls.
///
/// Loads stores, categories, and list items in parallel, then joins them
/// client-side into a Store → Category → Items hierarchy matching the web app.
///
/// Marked `@Observable` so SwiftUI views automatically re-render on any state
/// change.
@Observable
@MainActor
final class ShoppingListViewModel {

    // MARK: - State

    /// The list grouped into Store → Category → Items for display.
    private(set) var storeGroups: [StoreGroup] = []

    /// Total number of items on the list.
    private(set) var total: Int = 0

    /// Number of items already marked as done.
    private(set) var totalDone: Int = 0

    /// Full catalog of items used for local search suggestions.
    private(set) var allItems: [Item] = []

    /// `true` while the initial load (or a full refresh) is in flight.
    private(set) var isLoading: Bool = false

    /// `true` while a global mutation (finish shopping) is in flight.
    private(set) var isMutating: Bool = false

    /// The most recent error message to surface to the user.
    private(set) var errorMessage: String?

    /// The set of item IDs currently being mutated individually, used to
    /// disable row controls while their request is in flight.
    private(set) var mutatingItemIDs: Set<Int> = []

    // MARK: - Private raw state

    /// Flat list of all list items, kept in sync with `storeGroups`.
    private var items: [ListItem] = []

    /// Stores keyed by ID, populated on load.
    private var storesById: [Int: Store] = [:]

    /// Categories keyed by ID, populated on load.
    private var categoriesById: [Int: GroceriesAPI.Category] = [:]

    // MARK: - Dependencies

    private let apiClient: GroceriesAPIClient

    // MARK: - Init

    init(apiClient: GroceriesAPIClient) {
        self.apiClient = apiClient
    }

    // MARK: - Data loading

    /// Fetches stores, categories, and list items in parallel, then builds
    /// the grouped hierarchy.
    func load() async {
        guard !isLoading else { return }

        isLoading = true
        errorMessage = nil

        defer { isLoading = false }

        do {
            // All four calls are independent — fire them in parallel.
            async let storesFetch = apiClient.listStores()
            async let categoriesFetch = apiClient.listCategories()
            async let listFetch = apiClient.getList()
            async let itemsFetch = apiClient.listItems()

            let (stores, categories, list, fetchedItems) =
                try await (storesFetch, categoriesFetch, listFetch, itemsFetch)

            storesById = Dictionary(uniqueKeysWithValues: stores.map { ($0.id, $0) })
            categoriesById = Dictionary(
                uniqueKeysWithValues: categories.map { ($0.id, $0) }
                    as [(Int, GroceriesAPI.Category)])
            allItems = fetchedItems

            applyList(list)
        } catch {
            errorMessage = errorDescription(error)
        }
    }

    /// Silently refreshes only the list (stores/categories rarely change)
    /// without setting `isLoading`, so the UI doesn't flash during
    /// pull-to-refresh.
    func refresh() async {
        do {
            let list = try await apiClient.getList()
            applyList(list)
        } catch {
            errorMessage = errorDescription(error)
        }
    }

    // MARK: - Mutations

    /// Toggles the done state of a list item with an optimistic update.
    func toggleDone(for item: ListItem) async {
        guard !mutatingItemIDs.contains(item.itemID) else { return }

        let newDone = !item.done
        applyOptimisticUpdate(itemID: item.itemID, done: newDone)

        mutatingItemIDs.insert(item.itemID)
        defer { mutatingItemIDs.remove(item.itemID) }

        do {
            let updated = try await apiClient.updateListItem(
                itemID: item.itemID,
                done: newDone
            )
            reconcileItem(updated)
        } catch {
            // Roll back optimistic update.
            applyOptimisticUpdate(itemID: item.itemID, done: item.done)
            errorMessage = errorDescription(error)
        }
    }

    /// Updates the quantity string for a list item.
    func updateQuantity(for item: ListItem, quantity: String) async {
        guard !mutatingItemIDs.contains(item.itemID) else { return }

        mutatingItemIDs.insert(item.itemID)
        defer { mutatingItemIDs.remove(item.itemID) }

        do {
            let updated = try await apiClient.updateListItem(
                itemID: item.itemID,
                quantity: quantity
            )
            reconcileItem(updated)
        } catch {
            errorMessage = errorDescription(error)
        }
    }

    /// Removes an item from the list with an optimistic update.
    func remove(item: ListItem) async {
        guard !mutatingItemIDs.contains(item.itemID) else { return }

        let snapshot = items
        items.removeAll(where: { $0.itemID == item.itemID })
        rebuildGroups()

        mutatingItemIDs.insert(item.itemID)
        defer { mutatingItemIDs.remove(item.itemID) }

        do {
            try await apiClient.removeItemFromList(itemID: item.itemID)
        } catch {
            // Roll back.
            items = snapshot
            rebuildGroups()
            errorMessage = errorDescription(error)
        }
    }

    /// Finishes the shopping trip by removing all done items from the list.
    func finishShopping() async {
        guard !isMutating else { return }

        isMutating = true
        errorMessage = nil
        defer { isMutating = false }

        do {
            try await apiClient.finishShopping()
            let list = try await apiClient.getList()
            applyList(list)
        } catch {
            errorMessage = errorDescription(error)
        }
    }

    // MARK: - Computed helpers

    /// `true` when there is at least one item marked done.
    var hasDoneItems: Bool { totalDone > 0 }

    /// `true` when the list has no items at all.
    var isEmpty: Bool { items.isEmpty }

    /// Clears the currently displayed error message.
    func clearError() { errorMessage = nil }

    /// Returns items whose names case-insensitively start with `query`.
    /// An empty query (after trimming) returns no results.
    func searchItems(query: String) -> [Item] {
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return [] }

        return allItems.filter {
            $0.list == nil
                && $0.name.range(of: trimmed, options: [.caseInsensitive, .anchored]) != nil
        }
    }

    func addItem(itemID: Int?, name: String?, quantity: String) async throws {
        do {
            let listItem: ListItem

            if let itemID {
                listItem = try await apiClient.addItemToList(itemID: itemID, quantity: quantity)
            } else {
                let trimmedName = name?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
                guard !trimmedName.isEmpty else {
                    throw APIError.badRequest("name is required")
                }
                listItem = try await apiClient.addNewItemToList(name: trimmedName, quantity: quantity)
            }

            items.append(listItem)
            allItems.removeAll(where: { $0.id == listItem.itemID })
            total += 1
            totalDone = items.filter(\.done).count
            rebuildGroups()
        } catch {
            errorMessage = errorDescription(error)
            throw error
        }
    }

    // MARK: - Private helpers

    /// Applies a fresh `ShoppingList` response, updating totals and
    /// rebuilding the grouped hierarchy.
    private func applyList(_ list: ShoppingList) {
        items = list.items
        total = list.total
        totalDone = list.totalDone
        rebuildGroups()
    }

    /// Rebuilds `storeGroups` from the current `items`, `storesById`, and
    /// `categoriesById`. Mirrors the web app's Store → Category → Items
    /// grouping exactly.
    ///
    /// - Stores are shown in the order returned by the API.
    /// - Categories are shown in the order returned by the API, filtered to
    ///   those belonging to the current store.
    /// - Items with `category_id == 0` (uncategorized) are grouped under the
    ///   uncategorized store/category.
    private func rebuildGroups() {
        // Build a lookup from category_id → store_id using the cached data.
        // (category.storeID == 0 means the uncategorized store.)

        var groups: [StoreGroup] = []

        for store in storesById.values.sorted(by: { $0.name < $1.name }) {
            var categoryGroups: [CategoryGroup] = []

            let storeCategories: [GroceriesAPI.Category] = categoriesById.values
                .filter { $0.storeID == store.id }
                .sorted(by: { $0.name < $1.name })

            for category in storeCategories {
                let categoryItems = items.filter { $0.categoryID == category.id }
                guard !categoryItems.isEmpty else { continue }

                categoryGroups.append(
                    CategoryGroup(
                        id: category.id,
                        name: category.name,
                        items: categoryItems
                    ))
            }

            guard !categoryGroups.isEmpty else { continue }

            groups.append(
                StoreGroup(
                    id: store.id,
                    name: store.name,
                    categories: categoryGroups
                ))
        }

        storeGroups = groups
    }

    /// Optimistically mutates an item's `done` flag in the flat `items` array
    /// then rebuilds the grouped hierarchy.
    private func applyOptimisticUpdate(itemID: Int, done: Bool) {
        guard let index = items.firstIndex(where: { $0.itemID == itemID }) else { return }
        let existing = items[index]
        items[index] = ListItem(
            id: existing.id,
            itemID: existing.itemID,
            itemName: existing.itemName,
            categoryID: existing.categoryID,
            quantity: existing.quantity,
            done: done
        )
        totalDone = items.filter(\.done).count
        rebuildGroups()
    }

    /// Replaces the matching item in `items` with a server-confirmed value
    /// then rebuilds the grouped hierarchy.
    private func reconcileItem(_ updated: ListItem) {
        guard let index = items.firstIndex(where: { $0.itemID == updated.itemID }) else { return }
        items[index] = updated
        totalDone = items.filter(\.done).count
        rebuildGroups()
    }

    private func errorDescription(_ error: Error) -> String {
        if let apiError = error as? APIError {
            return apiError.errorDescription ?? error.localizedDescription
        }
        return error.localizedDescription
    }
}
