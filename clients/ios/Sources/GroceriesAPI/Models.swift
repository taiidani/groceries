import Foundation

// MARK: - Auth

public struct LoginRequest: Encodable, Sendable {
    public let username: String
    public let password: String

    public init(username: String, password: String) {
        self.username = username
        self.password = password
    }
}

public struct LoginResponse: Decodable, Sendable {
    public let token: String
    public let expiresAt: Date

    enum CodingKeys: String, CodingKey {
        case token
        case expiresAt = "expires_at"
    }
}

// MARK: - User

public struct User: Decodable, Identifiable, Sendable {
    public let id: Int
    public let name: String
    public let admin: Bool
}

// MARK: - Store

public struct Store: Decodable, Identifiable, Sendable {
    public let id: Int
    public let name: String
}

// MARK: - Category

public struct Category: Decodable, Identifiable, Sendable {
    public let id: Int
    public let storeID: Int
    public let name: String
    public let description: String
    public let itemCount: Int

    enum CodingKeys: String, CodingKey {
        case id
        case storeID = "store_id"
        case name
        case description
        case itemCount = "item_count"
    }
}

// MARK: - Item

public struct Item: Decodable, Identifiable, Sendable {
    public let id: Int
    public let categoryID: Int
    public let categoryName: String
    public let name: String
    /// Present when this item is currently on the shopping list.
    public let list: ListItemSummary?

    public init(id: Int, categoryID: Int, categoryName: String, name: String, list: ListItemSummary?) {
        self.id = id
        self.categoryID = categoryID
        self.categoryName = categoryName
        self.name = name
        self.list = list
    }

    enum CodingKeys: String, CodingKey {
        case id
        case categoryID = "category_id"
        case categoryName = "category_name"
        case name
        case list
    }
}

public struct CreateItemRequest: Encodable, Sendable {
    public let categoryID: Int
    public let name: String

    public init(categoryID: Int, name: String) {
        self.categoryID = categoryID
        self.name = name
    }

    enum CodingKeys: String, CodingKey {
        case categoryID = "category_id"
        case name
    }
}

public struct UpdateItemRequest: Encodable, Sendable {
    public let categoryID: Int
    public let name: String

    public init(categoryID: Int, name: String) {
        self.categoryID = categoryID
        self.name = name
    }

    enum CodingKeys: String, CodingKey {
        case categoryID = "category_id"
        case name
    }
}

// MARK: - Shopping List

/// A lightweight summary of a list entry embedded inside an `Item`.
public struct ListItemSummary: Decodable, Identifiable, Sendable {
    /// The list-entry ID (not the item ID).
    public let id: Int
    public let quantity: String
    public let done: Bool

    public init(id: Int, quantity: String, done: Bool) {
        self.id = id
        self.quantity = quantity
        self.done = done
    }
}

/// A full list entry as returned by `GET /api/v1/list`.
public struct ListItem: Decodable, Identifiable, Sendable {
    /// The list-entry ID.
    public let id: Int
    public let itemID: Int
    public let itemName: String
    public let categoryID: Int
    public let quantity: String
    public let done: Bool

    public init(
        id: Int,
        itemID: Int,
        itemName: String,
        categoryID: Int,
        quantity: String,
        done: Bool
    ) {
        self.id = id
        self.itemID = itemID
        self.itemName = itemName
        self.categoryID = categoryID
        self.quantity = quantity
        self.done = done
    }

    enum CodingKeys: String, CodingKey {
        case id
        case itemID = "item_id"
        case itemName = "item_name"
        case categoryID = "category_id"
        case quantity
        case done
    }
}

public struct ShoppingList: Decodable, Sendable {
    public let items: [ListItem]
    public let total: Int
    public let totalDone: Int

    enum CodingKeys: String, CodingKey {
        case items
        case total
        case totalDone = "total_done"
    }
}

public struct AddToListRequest: Encodable, Sendable {
    /// ID of an existing item to add.
    public let itemID: Int?
    /// Name of a new item to create and immediately add.
    public let name: String?
    public let quantity: String

    public init(itemID: Int? = nil, name: String? = nil, quantity: String = "") {
        self.itemID = itemID
        self.name = name
        self.quantity = quantity
    }

    enum CodingKeys: String, CodingKey {
        case itemID = "item_id"
        case name
        case quantity
    }
}

public struct UpdateListItemRequest: Encodable, Sendable {
    public let quantity: String?
    public let done: Bool?

    public init(quantity: String? = nil, done: Bool? = nil) {
        self.quantity = quantity
        self.done = done
    }
}

// MARK: - Recipe

public struct RecipeItem: Decodable, Identifiable, Sendable {
    public let id: Int
    public let recipeID: Int
    public let itemID: Int
    public let itemName: String
    public let quantity: String
    public let inList: Bool

    enum CodingKeys: String, CodingKey {
        case id
        case recipeID = "recipe_id"
        case itemID = "item_id"
        case itemName = "item_name"
        case quantity
        case inList = "in_list"
    }
}

public struct Recipe: Decodable, Identifiable, Sendable {
    public let id: Int
    public let name: String
    public let description: String
    public let createdAt: Date
    public let items: [RecipeItem]

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case description
        case createdAt = "created_at"
        case items
    }
}

public struct RecipeSummary: Decodable, Identifiable, Sendable {
    public let id: Int
    public let name: String
    public let description: String
    public let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case description
        case createdAt = "created_at"
    }
}

// MARK: - Errors

/// A structured error returned by the API.
public struct APIErrorResponse: Decodable, Sendable {
    public let error: String
}

/// Errors that can be thrown by `GroceriesAPIClient`.
public enum APIError: Error, LocalizedError, Sendable {
    case unauthorized
    case forbidden
    case notFound(String)
    case conflict(String)
    case badRequest(String)
    case serverError(String)
    case unexpectedStatus(Int)
    case decodingError(Error)

    public var errorDescription: String? {
        switch self {
        case .unauthorized:
            return "You are not logged in. Please sign in and try again."
        case .forbidden:
            return "You do not have permission to perform this action."
        case .notFound(let msg):
            return msg
        case .conflict(let msg):
            return msg
        case .badRequest(let msg):
            return msg
        case .serverError(let msg):
            return "Server error: \(msg)"
        case .unexpectedStatus(let code):
            return "Unexpected response from server (HTTP \(code))."
        case .decodingError(let err):
            return "Could not read server response: \(err.localizedDescription)"
        }
    }
}
