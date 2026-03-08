import Foundation

// MARK: - Store endpoints

extension GroceriesAPIClient {

    /// Returns all stores.
    public func listStores() async throws -> [Store] {
        let req = try request(method: "GET", path: "/api/v1/stores")
        return try await perform(req)
    }

    /// Returns all categories.
    public func listCategories() async throws -> [Category] {
        let req = try request(method: "GET", path: "/api/v1/categories")
        return try await perform(req)
    }
}
