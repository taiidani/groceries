import XCTest

@testable import GroceriesAPI

// MARK: - GroceriesAPIClientTests

private enum TestJSONHelper {
    static func jsonObject(from body: Data) throws -> [String: Any] {
        let object = try JSONSerialization.jsonObject(with: body)
        return try XCTUnwrap(object as? [String: Any])
    }
}

final class GroceriesAPIClientTests: XCTestCase {

    override func tearDown() {
        MockURLProtocol.clearRequestHandler()
        super.tearDown()
    }

    private func makeClient() -> GroceriesAPIClient {
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [MockURLProtocol.self]
        let session = URLSession(configuration: config)
        return GroceriesAPIClient(
            baseURL: URL(string: "http://localhost:3000")!,
            session: session
        )
    }

    // MARK: - JSON Decoding

    func testDecodeShoppingList() throws {
        let json = """
            {
                "items": [
                    {
                        "id": 1,
                        "item_id": 42,
                        "item_name": "Apples",
                        "category_id": 7,
                        "quantity": "6",
                        "done": false
                    },
                    {
                        "id": 2,
                        "item_id": 43,
                        "item_name": "Bread",
                        "category_id": 7,
                        "quantity": "",
                        "done": true
                    }
                ],
                "total": 2,
                "total_done": 1
            }
            """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let list = try JSONDecoder.apiDecoder.decode(ShoppingList.self, from: data)

        XCTAssertEqual(list.total, 2)
        XCTAssertEqual(list.totalDone, 1)
        XCTAssertEqual(list.items.count, 2)

        let first = list.items[0]
        XCTAssertEqual(first.id, 1)
        XCTAssertEqual(first.itemID, 42)
        XCTAssertEqual(first.itemName, "Apples")
        XCTAssertEqual(first.categoryID, 7)
        XCTAssertEqual(first.quantity, "6")
        XCTAssertFalse(first.done)

        let second = list.items[1]
        XCTAssertEqual(second.id, 2)
        XCTAssertEqual(second.itemName, "Bread")
        XCTAssertTrue(second.done)
    }

    func testDecodeListItemEmptyQuantity() throws {
        let json = """
            {
                "id": 5,
                "item_id": 99,
                "item_name": "Milk",
                "category_id": 3,
                "quantity": "",
                "done": false
            }
            """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let item = try JSONDecoder.apiDecoder.decode(ListItem.self, from: data)

        XCTAssertEqual(item.quantity, "")
    }

    func testDecodeLoginResponse() throws {
        let json = """
            {
                "token": "abc123",
                "expires_at": "2026-01-01T00:00:00Z"
            }
            """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let response = try JSONDecoder.apiDecoder.decode(LoginResponse.self, from: data)

        XCTAssertEqual(response.token, "abc123")
        XCTAssertNotNil(response.expiresAt)
    }

    func testDecodeUser() throws {
        let json = """
            {
                "id": 1,
                "name": "Alice",
                "admin": true
            }
            """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let user = try JSONDecoder.apiDecoder.decode(User.self, from: data)

        XCTAssertEqual(user.id, 1)
        XCTAssertEqual(user.name, "Alice")
        XCTAssertTrue(user.admin)
    }

    func testDecodeItemWithList() throws {
        let json = """
            {
                "id": 10,
                "category_id": 2,
                "category_name": "Produce",
                "name": "Bananas",
                "list": {
                    "id": 7,
                    "quantity": "4",
                    "done": false
                }
            }
            """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let item = try JSONDecoder.apiDecoder.decode(Item.self, from: data)

        XCTAssertEqual(item.id, 10)
        XCTAssertEqual(item.categoryName, "Produce")
        XCTAssertNotNil(item.list)
        XCTAssertEqual(item.list?.id, 7)
        XCTAssertEqual(item.list?.quantity, "4")
        XCTAssertFalse(item.list?.done ?? true)
    }

    func testDecodeItemWithoutList() throws {
        let json = """
            {
                "id": 11,
                "category_id": 2,
                "category_name": "Dairy",
                "name": "Cheese"
            }
            """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let item = try JSONDecoder.apiDecoder.decode(Item.self, from: data)

        XCTAssertNil(item.list)
    }

    func testDecodeAPIErrorResponse() throws {
        let json = """
            { "error": "item is already on the list" }
            """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let errorResponse = try JSONDecoder.apiDecoder.decode(APIErrorResponse.self, from: data)

        XCTAssertEqual(errorResponse.error, "item is already on the list")
    }

    // MARK: - Encoder round-trips

    func testEncodeAddToListRequestWithItemID() throws {
        let request = AddToListRequest(itemID: 42, quantity: "2 bags")
        let data = try JSONEncoder.apiEncoder.encode(request)
        let json = try XCTUnwrap(String(data: data, encoding: .utf8))

        XCTAssertTrue(json.contains("\"item_id\":42"))
        XCTAssertTrue(json.contains("\"quantity\":\"2 bags\""))
        // name should be omitted when nil
        XCTAssertFalse(json.contains("\"name\""))
    }

    func testEncodeAddToListRequestWithName() throws {
        let request = AddToListRequest(name: "Oat Milk", quantity: "")
        let data = try JSONEncoder.apiEncoder.encode(request)
        let json = try XCTUnwrap(String(data: data, encoding: .utf8))

        XCTAssertTrue(json.contains("\"name\":\"Oat Milk\""))
        XCTAssertTrue(json.contains("\"quantity\":\"\""))
    }

    func testEncodeUpdateListItemRequestPartial() throws {
        // Only done is set — quantity should be omitted.
        let request = UpdateListItemRequest(done: true)
        let data = try JSONEncoder.apiEncoder.encode(request)
        let json = try XCTUnwrap(String(data: data, encoding: .utf8))

        XCTAssertTrue(json.contains("\"done\":true"))
        XCTAssertFalse(json.contains("\"quantity\""))
    }

    func testEncodeCreateItemRequest() throws {
        let request = CreateItemRequest(categoryID: 7, name: "Apples")
        let data = try JSONEncoder.apiEncoder.encode(request)
        let json = try XCTUnwrap(String(data: data, encoding: .utf8))

        XCTAssertTrue(json.contains("\"category_id\":7"))
        XCTAssertTrue(json.contains("\"name\":\"Apples\""))
        XCTAssertFalse(json.contains("\"categoryID\""))
    }

    func testEncodeUpdateItemRequest() throws {
        let request = UpdateItemRequest(categoryID: 4, name: "Bread")
        let data = try JSONEncoder.apiEncoder.encode(request)
        let json = try XCTUnwrap(String(data: data, encoding: .utf8))

        XCTAssertTrue(json.contains("\"category_id\":4"))
        XCTAssertTrue(json.contains("\"name\":\"Bread\""))
        XCTAssertFalse(json.contains("\"categoryID\""))
    }

    // MARK: - APIError descriptions

    func testAPIErrorUnauthorizedDescription() {
        let error = APIError.unauthorized
        XCTAssertNotNil(error.errorDescription)
        XCTAssertFalse(error.errorDescription!.isEmpty)
    }

    func testAPIErrorNotFoundDescription() {
        let error = APIError.notFound("item not found")
        XCTAssertEqual(error.errorDescription, "item not found")
    }

    func testAPIErrorConflictDescription() {
        let error = APIError.conflict("item is already on the list")
        XCTAssertEqual(error.errorDescription, "item is already on the list")
    }

    func testAPIErrorBadRequestDescription() {
        let error = APIError.badRequest("username and password are required")
        XCTAssertEqual(error.errorDescription, "username and password are required")
    }

    func testAPIErrorServerErrorDescription() {
        let error = APIError.serverError("internal server error")
        XCTAssertTrue(error.errorDescription!.contains("internal server error"))
    }

    func testAPIErrorUnexpectedStatusDescription() {
        let error = APIError.unexpectedStatus(418)
        XCTAssertTrue(error.errorDescription!.contains("418"))
    }

    // MARK: - APIClient actor isolation

    func testClientInitiallyUnauthenticated() async {
        let client = GroceriesAPIClient(
            baseURL: URL(string: "http://localhost:3000")!
        )
        let authenticated = await client.isAuthenticated
        XCTAssertFalse(authenticated)
    }

    func testClientAuthenticatedWithToken() async {
        let client = GroceriesAPIClient(
            baseURL: URL(string: "http://localhost:3000")!,
            token: "some-token"
        )
        let authenticated = await client.isAuthenticated
        XCTAssertTrue(authenticated)
    }

    func testClientSetTokenNil() async {
        let client = GroceriesAPIClient(
            baseURL: URL(string: "http://localhost:3000")!,
            token: "some-token"
        )
        await client.setToken(nil)
        let authenticated = await client.isAuthenticated
        XCTAssertFalse(authenticated)
    }

    func testClientSetTokenValue() async {
        let client = GroceriesAPIClient(
            baseURL: URL(string: "http://localhost:3000")!
        )
        await client.setToken("new-token")
        let authenticated = await client.isAuthenticated
        XCTAssertTrue(authenticated)
    }

    func testMockURLProtocolRequestBodyData_readsFromHTTPBodyStream() throws {
        let expectedBody = Data("{\"name\":\"Oat Milk\"}".utf8)
        var request = URLRequest(url: URL(string: "http://localhost:3000/api/v1/list/items")!)
        request.httpMethod = "POST"
        request.httpBodyStream = InputStream(data: expectedBody)

        let extracted = MockURLProtocol.requestBodyData(from: request)

        XCTAssertEqual(extracted, expectedBody)
    }

    func testListItems_returnsAllItems() async throws {
        let responseJSON = """
            [
                {
                    "id": 1,
                    "category_id": 10,
                    "category_name": "Produce",
                    "name": "Apples"
                },
                {
                    "id": 2,
                    "category_id": 11,
                    "category_name": "Bakery",
                    "name": "Bread"
                }
            ]
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertEqual(request.url?.path, "/api/v1/items")

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()
        let items = try await client.listItems()

        XCTAssertEqual(items.count, 2)
        XCTAssertEqual(items[0].name, "Apples")
        XCTAssertEqual(items[1].name, "Bread")
    }

    func testListItems_withInListQueryEncodesQueryItem() async throws {
        let responseJSON = """
            [
                {
                    "id": 1,
                    "category_id": 10,
                    "category_name": "Produce",
                    "name": "Apples"
                }
            ]
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertEqual(request.url?.path, "/api/v1/items")

            let queryItems = URLComponents(url: try XCTUnwrap(request.url), resolvingAgainstBaseURL: false)?.queryItems
            XCTAssertEqual(queryItems?.count, 1)
            XCTAssertEqual(queryItems?.first?.name, "in_list")
            XCTAssertEqual(queryItems?.first?.value, "true")

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()
        let items = try await client.listItems(inList: true)

        XCTAssertEqual(items.count, 1)
        XCTAssertEqual(items[0].name, "Apples")
    }

    func testListItems_withCategoryIDQueryEncodesCategoryIDOnly() async throws {
        let responseJSON = """
            [
                {
                    "id": 3,
                    "category_id": 10,
                    "category_name": "Produce",
                    "name": "Carrots"
                }
            ]
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertEqual(request.url?.path, "/api/v1/items")

            let queryItems = URLComponents(url: try XCTUnwrap(request.url), resolvingAgainstBaseURL: false)?.queryItems
            XCTAssertEqual(queryItems?.count, 1)
            XCTAssertEqual(queryItems?.first?.name, "category_id")
            XCTAssertEqual(queryItems?.first?.value, "10")

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()
        let items = try await client.listItems(categoryID: 10)

        XCTAssertEqual(items.count, 1)
        XCTAssertEqual(items[0].name, "Carrots")
    }

    func testListItems_withInListFalseQueryEncodesFalse() async throws {
        let responseJSON = """
            [
                {
                    "id": 4,
                    "category_id": 12,
                    "category_name": "Pantry",
                    "name": "Pasta"
                }
            ]
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertEqual(request.url?.path, "/api/v1/items")

            let queryItems = URLComponents(url: try XCTUnwrap(request.url), resolvingAgainstBaseURL: false)?.queryItems
            XCTAssertEqual(queryItems?.count, 1)
            XCTAssertEqual(queryItems?.first?.name, "in_list")
            XCTAssertEqual(queryItems?.first?.value, "false")

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()
        let items = try await client.listItems(inList: false)

        XCTAssertEqual(items.count, 1)
        XCTAssertEqual(items[0].name, "Pasta")
    }

    func testListItems_withCategoryIDAndInListQueryEncodesBothItems() async throws {
        let responseJSON = """
            [
                {
                    "id": 5,
                    "category_id": 2,
                    "category_name": "Dairy",
                    "name": "Yogurt"
                }
            ]
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "GET")
            XCTAssertEqual(request.url?.path, "/api/v1/items")

            let queryItems = URLComponents(url: try XCTUnwrap(request.url), resolvingAgainstBaseURL: false)?.queryItems ?? []
            XCTAssertEqual(queryItems.count, 2)

            let queryByName = Dictionary(uniqueKeysWithValues: queryItems.map { ($0.name, $0.value) })
            XCTAssertEqual(queryByName["category_id"], "2")
            XCTAssertEqual(queryByName["in_list"], "false")

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()
        let items = try await client.listItems(categoryID: 2, inList: false)

        XCTAssertEqual(items.count, 1)
        XCTAssertEqual(items[0].name, "Yogurt")
    }

    func testCreateItem_postsExpectedBody() async throws {
        let responseJSON = """
            {
                "id": 21,
                "category_id": 4,
                "category_name": "Bakery",
                "name": "Bagels"
            }
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.url?.path, "/api/v1/items")

            let body = try XCTUnwrap(MockURLProtocol.requestBodyData(from: request))
            let payload = try TestJSONHelper.jsonObject(from: body)
            XCTAssertEqual(payload["category_id"] as? Int, 4)
            XCTAssertEqual(payload["name"] as? String, "Bagels")
            XCTAssertNil(payload["categoryID"])

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()
        let item = try await client.createItem(categoryID: 4, name: "Bagels")

        XCTAssertEqual(item.id, 21)
        XCTAssertEqual(item.name, "Bagels")
        XCTAssertEqual(item.categoryID, 4)
    }

    func testUpdateItem_putsExpectedBody() async throws {
        let responseJSON = """
            {
                "id": 21,
                "category_id": 5,
                "category_name": "Pantry",
                "name": "Granola"
            }
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "PUT")
            XCTAssertEqual(request.url?.path, "/api/v1/items/21")

            let body = try XCTUnwrap(MockURLProtocol.requestBodyData(from: request))
            let payload = try TestJSONHelper.jsonObject(from: body)
            XCTAssertEqual(payload["category_id"] as? Int, 5)
            XCTAssertEqual(payload["name"] as? String, "Granola")
            XCTAssertNil(payload["categoryID"])

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()
        let item = try await client.updateItem(id: 21, categoryID: 5, name: "Granola")

        XCTAssertEqual(item.id, 21)
        XCTAssertEqual(item.name, "Granola")
        XCTAssertEqual(item.categoryID, 5)
    }

    func testDeleteItem_sendsDelete() async throws {
        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "DELETE")
            XCTAssertEqual(request.url?.path, "/api/v1/items/21")

            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 204,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, Data())
        }

        let client = makeClient()
        try await client.deleteItem(id: 21)
    }

    func testDeleteItem_conflict_throws409() async throws {
        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "DELETE")
            XCTAssertEqual(request.url?.path, "/api/v1/items/21")

            let data = try XCTUnwrap("{\"error\":\"item is on the shopping list\"}".data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 409,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()

        do {
            try await client.deleteItem(id: 21)
            XCTFail("Expected conflict error")
        } catch let error as APIError {
            switch error {
            case .conflict(let message):
                XCTAssertEqual(message, "item is on the shopping list")
            default:
                XCTFail("Expected APIError.conflict, got \(error)")
            }
        }
    }

    func testUpdateItem_notFound_throws404() async throws {
        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "PUT")
            XCTAssertEqual(request.url?.path, "/api/v1/items/999")

            let body = try XCTUnwrap(MockURLProtocol.requestBodyData(from: request))
            let payload = try TestJSONHelper.jsonObject(from: body)
            XCTAssertEqual(payload["category_id"] as? Int, 5)
            XCTAssertEqual(payload["name"] as? String, "Granola")

            let data = try XCTUnwrap("{\"error\":\"item not found\"}".data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 404,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()

        do {
            _ = try await client.updateItem(id: 999, categoryID: 5, name: "Granola")
            XCTFail("Expected notFound error")
        } catch let error as APIError {
            switch error {
            case .notFound(let message):
                XCTAssertEqual(message, "item not found")
            default:
                XCTFail("Expected APIError.notFound, got \(error)")
            }
        }
    }

    func testDeleteItem_notFound_throws404() async throws {
        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "DELETE")
            XCTAssertEqual(request.url?.path, "/api/v1/items/999")

            let data = try XCTUnwrap("{\"error\":\"item not found\"}".data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 404,
                    httpVersion: nil,
                    headerFields: nil
                )
            )

            return (response, data)
        }

        let client = makeClient()

        do {
            try await client.deleteItem(id: 999)
            XCTFail("Expected notFound error")
        } catch let error as APIError {
            switch error {
            case .notFound(let message):
                XCTAssertEqual(message, "item not found")
            default:
                XCTFail("Expected APIError.notFound, got \(error)")
            }
        }
    }

    func testAddItemToList_existingItem() async throws {
        let responseJSON = """
            {
                "id": 12,
                "item_id": 42,
                "item_name": "Apples",
                "category_id": 10,
                "quantity": "2",
                "done": false
            }
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.url?.path, "/api/v1/list/items")

            let body = try XCTUnwrap(MockURLProtocol.requestBodyData(from: request))
            let bodyString = try XCTUnwrap(String(data: body, encoding: .utf8))
            XCTAssertTrue(bodyString.contains("\"item_id\":42"))
            XCTAssertTrue(bodyString.contains("\"quantity\":\"2\""))
            XCTAssertFalse(bodyString.contains("\"name\""))

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )
            return (response, data)
        }

        let client = makeClient()
        let item = try await client.addItemToList(itemID: 42, quantity: "2")

        XCTAssertEqual(item.id, 12)
        XCTAssertEqual(item.itemID, 42)
        XCTAssertEqual(item.itemName, "Apples")
        XCTAssertEqual(item.quantity, "2")
        XCTAssertFalse(item.done)
    }

    func testAddNewItemToList_freeText() async throws {
        let responseJSON = """
            {
                "id": 17,
                "item_id": 88,
                "item_name": "Oat Milk",
                "category_id": 3,
                "quantity": "1",
                "done": false
            }
            """

        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.url?.path, "/api/v1/list/items")

            let body = try XCTUnwrap(MockURLProtocol.requestBodyData(from: request))
            let bodyString = try XCTUnwrap(String(data: body, encoding: .utf8))
            XCTAssertTrue(bodyString.contains("\"name\":\"Oat Milk\""))
            XCTAssertTrue(bodyString.contains("\"quantity\":\"1\""))
            XCTAssertFalse(bodyString.contains("\"item_id\""))

            let data = try XCTUnwrap(responseJSON.data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 200,
                    httpVersion: nil,
                    headerFields: nil
                )
            )
            return (response, data)
        }

        let client = makeClient()
        let item = try await client.addNewItemToList(name: "Oat Milk", quantity: "1")

        XCTAssertEqual(item.id, 17)
        XCTAssertEqual(item.itemID, 88)
        XCTAssertEqual(item.itemName, "Oat Milk")
        XCTAssertEqual(item.quantity, "1")
        XCTAssertFalse(item.done)
    }

    func testAddToList_conflict_throws409() async throws {
        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.url?.path, "/api/v1/list/items")

            let body = try XCTUnwrap(MockURLProtocol.requestBodyData(from: request))
            let bodyString = try XCTUnwrap(String(data: body, encoding: .utf8))
            XCTAssertTrue(bodyString.contains("\"item_id\":42"))
            XCTAssertTrue(bodyString.contains("\"quantity\":\"\""))

            let data = try XCTUnwrap("{\"error\":\"item is already on the list\"}".data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 409,
                    httpVersion: nil,
                    headerFields: nil
                )
            )
            return (response, data)
        }

        let client = makeClient()

        do {
            _ = try await client.addItemToList(itemID: 42)
            XCTFail("Expected conflict error")
        } catch let error as APIError {
            switch error {
            case .conflict(let message):
                XCTAssertEqual(message, "item is already on the list")
            default:
                XCTFail("Expected APIError.conflict, got \(error)")
            }
        }
    }

    func testAddToList_notFound_throws404() async throws {
        MockURLProtocol.setRequestHandler { request in
            XCTAssertEqual(request.httpMethod, "POST")
            XCTAssertEqual(request.url?.path, "/api/v1/list/items")

            let body = try XCTUnwrap(MockURLProtocol.requestBodyData(from: request))
            let bodyString = try XCTUnwrap(String(data: body, encoding: .utf8))
            XCTAssertTrue(bodyString.contains("\"item_id\":999"))
            XCTAssertTrue(bodyString.contains("\"quantity\":\"\""))

            let data = try XCTUnwrap("{\"error\":\"item not found\"}".data(using: .utf8))
            let response = try XCTUnwrap(
                HTTPURLResponse(
                    url: try XCTUnwrap(request.url),
                    statusCode: 404,
                    httpVersion: nil,
                    headerFields: nil
                )
            )
            return (response, data)
        }

        let client = makeClient()

        do {
            _ = try await client.addItemToList(itemID: 999)
            XCTFail("Expected notFound error")
        } catch let error as APIError {
            switch error {
            case .notFound(let message):
                XCTAssertEqual(message, "item not found")
            default:
                XCTFail("Expected APIError.notFound, got \(error)")
            }
        }
    }
}
