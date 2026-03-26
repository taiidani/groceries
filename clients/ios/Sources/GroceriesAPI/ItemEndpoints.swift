import Foundation

// MARK: - Item endpoints

extension GroceriesAPIClient {

    /// Returns all items, optionally filtered by category and list membership.
    public func listItems(categoryID: Int? = nil, inList: Bool? = nil) async throws -> [Item] {
        var queryItems: [URLQueryItem] = []

        if let categoryID {
            queryItems.append(URLQueryItem(name: "category_id", value: String(categoryID)))
        }

        if let inList {
            queryItems.append(URLQueryItem(name: "in_list", value: inList ? "true" : "false"))
        }

        let req = try request(
            method: "GET",
            path: "/api/v1/items",
            queryItems: queryItems.isEmpty ? nil : queryItems
        )
        return try await perform(req)
    }

    /// Creates a new grocery item.
    public func createItem(categoryID: Int, name: String) async throws -> Item {
        let body = CreateItemRequest(categoryID: categoryID, name: name)
        let req = try request(method: "POST", path: "/api/v1/items", body: body)
        return try await perform(req)
    }

    /// Updates an existing grocery item.
    public func updateItem(id: Int, categoryID: Int, name: String) async throws -> Item {
        let body = UpdateItemRequest(categoryID: categoryID, name: name)
        let req = try request(method: "PUT", path: "/api/v1/items/\(id)", body: body)
        return try await perform(req)
    }

    /// Deletes a grocery item.
    public func deleteItem(id: Int) async throws {
        let req = try request(method: "DELETE", path: "/api/v1/items/\(id)")
        try await performVoid(req)
    }
}
