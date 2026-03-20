import Foundation

// MARK: - Shopping list endpoints

extension GroceriesAPIClient {

    /// Returns all items that can be added to the shopping list.
    public func listItems() async throws -> [Item] {
        let req = try request(method: "GET", path: "/api/v1/items")
        return try await perform(req)
    }

    /// Returns the current shopping list with totals.
    public func getList() async throws -> ShoppingList {
        let req = try request(method: "GET", path: "/api/v1/list")
        return try await perform(req)
    }

    /// Adds an existing item (by ID) to the shopping list.
    public func addItemToList(itemID: Int, quantity: String = "") async throws -> ListItem {
        let body = AddToListRequest(itemID: itemID, quantity: quantity)
        let req = try request(method: "POST", path: "/api/v1/list/items", body: body)
        return try await perform(req)
    }

    /// Creates a new item by name and immediately adds it to the shopping list.
    public func addNewItemToList(name: String, quantity: String = "") async throws -> ListItem {
        let body = AddToListRequest(name: name, quantity: quantity)
        let req = try request(method: "POST", path: "/api/v1/list/items", body: body)
        return try await perform(req)
    }

    /// Updates the quantity and/or done status of a list entry.
    ///
    /// - Parameter itemID: The grocery item ID (not the list-entry ID).
    public func updateListItem(
        itemID: Int,
        quantity: String? = nil,
        done: Bool? = nil
    ) async throws -> ListItem {
        let body = UpdateListItemRequest(quantity: quantity, done: done)
        let req = try request(method: "PUT", path: "/api/v1/list/items/\(itemID)", body: body)
        return try await perform(req)
    }

    /// Removes an item from the shopping list.
    ///
    /// - Parameter itemID: The grocery item ID (not the list-entry ID).
    public func removeItemFromList(itemID: Int) async throws {
        let req = try request(method: "DELETE", path: "/api/v1/list/items/\(itemID)")
        try await performVoid(req)
    }

    /// Clears all done items from the list, finishing the shopping trip.
    public func finishShopping() async throws {
        let req = try request(method: "POST", path: "/api/v1/list/finish")
        try await performVoid(req)
    }
}
