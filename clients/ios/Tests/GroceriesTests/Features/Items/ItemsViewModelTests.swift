import GroceriesAPI
import XCTest

@testable import Aisle4

@MainActor
final class ItemsViewModelTests: XCTestCase {
    func test_filteredItems_matchesCaseInsensitiveSubstring() async throws {
        let api = MockItemsAPI(
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 1,
                    "store_id": 1,
                    "name": "Dairy",
                    "description": "",
                    "item_count": 3
                  }
                ]
                """
            ),
            items: try decodeItems(
                """
                [
                  {
                    "id": 1,
                    "category_id": 1,
                    "category_name": "Dairy",
                    "name": "Almond Milk"
                  },
                  {
                    "id": 2,
                    "category_id": 1,
                    "category_name": "Dairy",
                    "name": "Whole Milk"
                  },
                  {
                    "id": 3,
                    "category_id": 1,
                    "category_name": "Dairy",
                    "name": "Bread"
                  }
                ]
                """
            )
        )
        let viewModel = ItemsViewModel(api: api)

        await viewModel.load()
        viewModel.searchText = "mIlK"

        XCTAssertEqual(viewModel.filteredItems.map(\.name), ["Almond Milk", "Whole Milk"])
    }

    func test_filteredItems_composesInListOnlyAndSearch() async throws {
        let api = MockItemsAPI(
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 1,
                    "store_id": 1,
                    "name": "Dairy",
                    "description": "",
                    "item_count": 3
                  }
                ]
                """
            ),
            items: try decodeItems(
                """
                [
                  {
                    "id": 1,
                    "category_id": 1,
                    "category_name": "Dairy",
                    "name": "Almond Milk",
                    "list": { "id": 100, "quantity": "1", "done": false }
                  },
                  {
                    "id": 2,
                    "category_id": 1,
                    "category_name": "Dairy",
                    "name": "Coconut Milk"
                  },
                  {
                    "id": 3,
                    "category_id": 1,
                    "category_name": "Dairy",
                    "name": "Pasta",
                    "list": { "id": 101, "quantity": "1", "done": false }
                  }
                ]
                """
            )
        )
        let viewModel = ItemsViewModel(api: api)

        await viewModel.load()
        viewModel.inListOnly = true
        viewModel.searchText = "milk"

        XCTAssertEqual(viewModel.filteredItems.map(\.name), ["Almond Milk"])
    }

    func test_addAndUpdate_rejectTrimmedEmptyName() async throws {
        let api = MockItemsAPI(
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 1,
                    "store_id": 1,
                    "name": "Dairy",
                    "description": "",
                    "item_count": 0
                  }
                ]
                """
            ),
            items: []
        )
        let viewModel = ItemsViewModel(api: api)

        let addResult = await viewModel.addItem(name: "   ", categoryID: 1)
        let updateResult = await viewModel.updateItem(id: 42, name: "\n\t", categoryID: 1)

        XCTAssertFalse(addResult)
        XCTAssertFalse(updateResult)
        XCTAssertEqual(api.createItemCallCount, 0)
        XCTAssertEqual(api.updateItemCallCount, 0)
        XCTAssertEqual(viewModel.errorMessage, "Item name is required.")
    }

    func test_addAndUpdate_requireCategory() async throws {
        let api = MockItemsAPI(
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 1,
                    "store_id": 1,
                    "name": "Dairy",
                    "description": "",
                    "item_count": 0
                  }
                ]
                """
            ),
            items: []
        )
        let viewModel = ItemsViewModel(api: api)

        let addResult = await viewModel.addItem(name: "Apples", categoryID: nil)
        let updateResult = await viewModel.updateItem(id: 42, name: "Apples", categoryID: nil)

        XCTAssertFalse(addResult)
        XCTAssertFalse(updateResult)
        XCTAssertEqual(api.createItemCallCount, 0)
        XCTAssertEqual(api.updateItemCallCount, 0)
        XCTAssertEqual(viewModel.errorMessage, "Category is required.")
    }

    func test_loadFailure_keepsEmptySafeState_andRetryPathRecovers() async throws {
        let api = MockItemsAPI(
            categories: [],
            items: [],
            listCategoriesError: APIError.serverError("boom"),
            listItemsError: APIError.serverError("boom")
        )
        let viewModel = ItemsViewModel(api: api)

        await viewModel.load()

        XCTAssertTrue(viewModel.items.isEmpty)
        XCTAssertTrue(viewModel.filteredItems.isEmpty)
        XCTAssertTrue(viewModel.categories.isEmpty)
        XCTAssertNotNil(viewModel.errorMessage)

        api.listCategoriesError = nil
        api.listItemsError = nil
        api.categories = try decodeCategories(
            """
            [
              {
                "id": 1,
                "store_id": 1,
                "name": "Dairy",
                "description": "",
                "item_count": 0
              }
            ]
            """
        )
        api.items = try decodeItems(
            """
            [
              {
                "id": 1,
                "category_id": 1,
                "category_name": "Dairy",
                "name": "Milk"
              }
            ]
            """
        )

        await viewModel.retryLoad()

        XCTAssertEqual(viewModel.items.map(\.name), ["Milk"])
        XCTAssertEqual(viewModel.categories.map(\.name), ["Dairy"])
        XCTAssertNil(viewModel.errorMessage)
        XCTAssertEqual(api.listCategoriesCallCount, 2)
        XCTAssertEqual(api.listItemsCallCount, 2)
    }

    func test_addUpdateDeleteSuccess_updateLocalCache() async throws {
        let api = MockItemsAPI(
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 1,
                    "store_id": 1,
                    "name": "Dairy",
                    "description": "",
                    "item_count": 0
                  }
                ]
                """
            ),
            items: []
        )
        api.createItemResult = try decodeItem(
            """
            {
              "id": 10,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Oat Milk"
            }
            """
        )
        api.updateItemResult = try decodeItem(
            """
            {
              "id": 10,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Oat Milk Unsweetened"
            }
            """
        )

        let viewModel = ItemsViewModel(api: api)

        let addOK = await viewModel.addItem(name: "Oat Milk", categoryID: 1)
        XCTAssertTrue(addOK)
        XCTAssertEqual(viewModel.items.map(\.name), ["Oat Milk"])

        let updateOK = await viewModel.updateItem(id: 10, name: "Oat Milk Unsweetened", categoryID: 1)
        XCTAssertTrue(updateOK)
        XCTAssertEqual(viewModel.items.map(\.name), ["Oat Milk Unsweetened"])

        let deleteOK = await viewModel.deleteItem(id: 10)
        XCTAssertTrue(deleteOK)
        XCTAssertTrue(viewModel.items.isEmpty)
    }

    func test_setInListSuccess_postsMembershipNotificationPayloadTypes() async throws {
        let notificationCenter = NotificationCenter()
        let api = MockItemsAPI(
            categories: [],
            items: [
                try decodeItem(
                    """
                    {
                      "id": 1,
                      "category_id": 1,
                      "category_name": "Dairy",
                      "name": "Milk"
                    }
                    """
                )
            ]
        )
        api.listItemsAfterSetInList = [
            try decodeItem(
                """
                {
                  "id": 1,
                  "category_id": 1,
                  "category_name": "Dairy",
                  "name": "Milk",
                  "list": { "id": 22, "quantity": "1", "done": false }
                }
                """
            )
        ]

        let viewModel = ItemsViewModel(api: api, notificationCenter: notificationCenter)
        await viewModel.load()

        let recorder = NotificationRecorder()
        let token = notificationCenter.addObserver(
            forName: AppEvents.MembershipChanged.name,
            object: nil,
            queue: nil
        ) { note in
            recorder.record(note)
        }
        defer { notificationCenter.removeObserver(token) }

        let ok = await viewModel.setInList(itemID: 1, isInList: true)

        XCTAssertTrue(ok)
        let userInfo = try XCTUnwrap(recorder.lastUserInfo)
        XCTAssertTrue(userInfo[AppEvents.MembershipChanged.itemIDKey] is Int)
        XCTAssertTrue(userInfo[AppEvents.MembershipChanged.isInListKey] is Bool)
        XCTAssertTrue(userInfo[AppEvents.MembershipChanged.changedAtKey] is Date)
    }

    func test_toggleAndDeleteFailure_preserveState_andToggleFailureDoesNotPostNotification() async throws {
        let notificationCenter = NotificationCenter()
        let original = try decodeItem(
            """
            {
              "id": 1,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )
        let api = MockItemsAPI(categories: [], items: [original])
        api.addItemToListError = APIError.serverError("toggle fail")
        api.deleteItemError = APIError.serverError("delete fail")

        let viewModel = ItemsViewModel(api: api, notificationCenter: notificationCenter)
        await viewModel.load()

        let recorder = NotificationRecorder()
        let token = notificationCenter.addObserver(
            forName: AppEvents.MembershipChanged.name,
            object: nil,
            queue: nil
        ) { _ in
            recorder.recordCount()
        }
        defer { notificationCenter.removeObserver(token) }

        let toggleOK = await viewModel.setInList(itemID: 1, isInList: true)
        let deleteOK = await viewModel.deleteItem(id: 1)

        XCTAssertFalse(toggleOK)
        XCTAssertFalse(deleteOK)
        XCTAssertEqual(recorder.count, 0)
        XCTAssertEqual(viewModel.items.map(\.id), [1])
    }

    func test_duplicateSubmitProtection_allowsSingleInFlightCallPerOperation() async throws {
        let api = MockItemsAPI(categories: [], items: [])
        api.createItemResult = try decodeItem(
            """
            {
              "id": 1,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )
        api.waitForCreate = true

        let viewModel = ItemsViewModel(api: api)

        let first = Task { await viewModel.addItem(name: "Milk", categoryID: 1) }
        let second = Task { await viewModel.addItem(name: "Milk", categoryID: 1) }

        await Task.yield()
        api.releaseCreateContinuation()

        _ = await first.value
        _ = await second.value

        XCTAssertEqual(api.createItemCallCount, 1)
    }

    func test_retryAndRefresh_refetchAndReplaceStaleCache() async throws {
        let api = MockItemsAPI(
            categories: [
                try decodeCategory(
                    """
                    {
                      "id": 1,
                      "store_id": 1,
                      "name": "Old",
                      "description": "",
                      "item_count": 0
                    }
                    """
                )
            ],
            items: [
                try decodeItem(
                    """
                    {
                      "id": 1,
                      "category_id": 1,
                      "category_name": "Old",
                      "name": "Old Item"
                    }
                    """
                )
            ]
        )
        let viewModel = ItemsViewModel(api: api)

        await viewModel.load()
        XCTAssertEqual(viewModel.items.map(\.name), ["Old Item"])

        api.categories = [
            try decodeCategory(
                """
                {
                  "id": 2,
                  "store_id": 1,
                  "name": "New",
                  "description": "",
                  "item_count": 0
                }
                """
            )
        ]
        api.items = [
            try decodeItem(
                """
                {
                  "id": 2,
                  "category_id": 2,
                  "category_name": "New",
                  "name": "New Item"
                }
                """
            )
        ]

        await viewModel.retryLoad()
        XCTAssertEqual(viewModel.items.map(\.name), ["New Item"])
        XCTAssertEqual(viewModel.categories.map(\.name), ["New"])

        api.items = [
            try decodeItem(
                """
                {
                  "id": 3,
                  "category_id": 2,
                  "category_name": "New",
                  "name": "Newest Item"
                }
                """
            )
        ]

        await viewModel.refresh()
        XCTAssertEqual(viewModel.items.map(\.name), ["Newest Item"])
        XCTAssertEqual(api.listCategoriesCallCount, 3)
        XCTAssertEqual(api.listItemsCallCount, 3)
    }
}

private func decodeCategories(_ json: String) throws -> [GroceriesAPI.Category] {
    let data = try XCTUnwrap(json.data(using: .utf8))
    return try JSONDecoder().decode([GroceriesAPI.Category].self, from: data)
}

private func decodeItems(_ json: String) throws -> [Item] {
    let data = try XCTUnwrap(json.data(using: .utf8))
    return try JSONDecoder().decode([Item].self, from: data)
}

private func decodeCategory(_ json: String) throws -> GroceriesAPI.Category {
    let data = try XCTUnwrap(json.data(using: .utf8))
    return try JSONDecoder().decode(GroceriesAPI.Category.self, from: data)
}

private func decodeItem(_ json: String) throws -> Item {
    let data = try XCTUnwrap(json.data(using: .utf8))
    return try JSONDecoder().decode(Item.self, from: data)
}

private final class NotificationRecorder: @unchecked Sendable {
    private let lock = NSLock()
    private var storedUserInfo: [AnyHashable: Any]?
    private var storedCount = 0

    var lastUserInfo: [AnyHashable: Any]? {
        lock.lock()
        defer { lock.unlock() }
        return storedUserInfo
    }

    var count: Int {
        lock.lock()
        defer { lock.unlock() }
        return storedCount
    }

    func record(_ notification: Notification) {
        lock.lock()
        storedUserInfo = notification.userInfo
        storedCount += 1
        lock.unlock()
    }

    func recordCount() {
        lock.lock()
        storedCount += 1
        lock.unlock()
    }
}

@MainActor
private final class MockItemsAPI: ItemsAPI {
    var categories: [GroceriesAPI.Category]
    var items: [Item]

    var listCategoriesError: Error?
    var listItemsError: Error?
    var createItemError: Error?
    var updateItemError: Error?
    var deleteItemError: Error?
    var addItemToListError: Error?
    var removeItemFromListError: Error?

    var createItemResult: Item?
    var updateItemResult: Item?
    var listItemsAfterSetInList: [Item]?

    var waitForCreate = false
    private var createContinuation: CheckedContinuation<Void, Never>?

    private(set) var createItemCallCount = 0
    private(set) var updateItemCallCount = 0
    private(set) var listCategoriesCallCount = 0
    private(set) var listItemsCallCount = 0
    private(set) var deleteItemCallCount = 0
    private(set) var addItemToListCallCount = 0
    private(set) var removeItemFromListCallCount = 0

    init(categories: [GroceriesAPI.Category], items: [Item], listCategoriesError: Error? = nil, listItemsError: Error? = nil) {
        self.categories = categories
        self.items = items
        self.listCategoriesError = listCategoriesError
        self.listItemsError = listItemsError
    }

    func listCategories() async throws -> [GroceriesAPI.Category] {
        listCategoriesCallCount += 1
        if let listCategoriesError {
            throw listCategoriesError
        }
        return categories
    }

    func listItems(inList: Bool?) async throws -> [Item] {
        listItemsCallCount += 1
        if let listItemsError {
            throw listItemsError
        }

        if let listItemsAfterSetInList {
            return listItemsAfterSetInList
        }

        if let inList {
            return items.filter { ($0.list != nil) == inList }
        }
        return items
    }

    func createItem(categoryID: Int, name: String) async throws -> Item {
        createItemCallCount += 1

        if waitForCreate {
            await withCheckedContinuation { continuation in
                createContinuation = continuation
            }
        }

        if let createItemError {
            throw createItemError
        }

        let result = try XCTUnwrap(createItemResult)
        items.append(result)
        return result
    }

    func updateItem(id: Int, categoryID: Int, name: String) async throws -> Item {
        updateItemCallCount += 1

        if let updateItemError {
            throw updateItemError
        }

        let result = try XCTUnwrap(updateItemResult)
        if let index = items.firstIndex(where: { $0.id == id }) {
            items[index] = result
        }
        return result
    }

    func deleteItem(id: Int) async throws {
        deleteItemCallCount += 1
        if let deleteItemError {
            throw deleteItemError
        }
        items.removeAll(where: { $0.id == id })
    }

    func addItemToList(itemID: Int) async throws -> Item {
        addItemToListCallCount += 1
        if let addItemToListError {
            throw addItemToListError
        }

        if let listItemsAfterSetInList,
            let item = listItemsAfterSetInList.first(where: { $0.id == itemID })
        {
            items = listItemsAfterSetInList
            return item
        }

        return try XCTUnwrap(items.first(where: { $0.id == itemID }))
    }

    func removeItemFromList(itemID: Int) async throws {
        removeItemFromListCallCount += 1
        if let removeItemFromListError {
            throw removeItemFromListError
        }

        if let listItemsAfterSetInList {
            items = listItemsAfterSetInList
        }
    }

    func releaseCreateContinuation() {
        createContinuation?.resume()
        createContinuation = nil
    }
}
