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

    func test_load_forceOverlap_keepsNewestLoadResults() async throws {
        let staleCategory = try decodeCategory(
            """
            {
              "id": 1,
              "store_id": 1,
              "name": "Stale",
              "description": "",
              "item_count": 1
            }
            """
        )
        let freshCategory = try decodeCategory(
            """
            {
              "id": 2,
              "store_id": 1,
              "name": "Fresh",
              "description": "",
              "item_count": 1
            }
            """
        )
        let staleItem = try decodeItem(
            """
            {
              "id": 1,
              "category_id": 1,
              "category_name": "Stale",
              "name": "Stale Item"
            }
            """
        )
        let freshItem = try decodeItem(
            """
            {
              "id": 2,
              "category_id": 2,
              "category_name": "Fresh",
              "name": "Fresh Item"
            }
            """
        )

        let api = MockItemsAPI(categories: [staleCategory], items: [staleItem])
        api.listItemsResponsesByCall = [
            1: [staleItem],
            2: [freshItem],
        ]
        api.blockedListItemsCallIndices = [1]

        let parked = expectation(description: "first listItems call parked")
        api.onListItemsCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let firstLoad = Task { await viewModel.load() }
        await fulfillment(of: [parked], timeout: 1.0)

        api.categories = [freshCategory]
        let forcedRefresh = Task { await viewModel.retryLoad() }
        await forcedRefresh.value

        api.releaseListItemsCall(1)
        await firstLoad.value

        XCTAssertEqual(viewModel.items.map(\.name), ["Fresh Item"])
        XCTAssertEqual(viewModel.categories.map(\.name), ["Fresh"])
    }

    func test_setInListSuccess_postsMembershipNotificationContract() async throws {
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
        XCTAssertEqual(recorder.count, 1)

        let notification = try XCTUnwrap(recorder.lastNotification)
        XCTAssertEqual(notification.name, AppEvents.MembershipChanged.name)
        XCTAssertNil(notification.object)

        let userInfo = try XCTUnwrap(notification.userInfo)
        XCTAssertEqual(userInfo[AppEvents.MembershipChanged.itemIDKey] as? Int, 1)
        XCTAssertEqual(userInfo[AppEvents.MembershipChanged.isInListKey] as? Bool, true)
        XCTAssertNotNil(userInfo[AppEvents.MembershipChanged.changedAtKey] as? Date)
    }

    func test_setInListToggleOffSuccess_callsRemoveRefreshesAndPostsFalseNotification() async throws {
        let notificationCenter = NotificationCenter()
        let original = try decodeItem(
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
        let toggledOff = try decodeItem(
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
        api.listItemsAfterSetInList = [toggledOff]

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

        let ok = await viewModel.setInList(itemID: 1, isInList: false)

        XCTAssertTrue(ok)
        XCTAssertEqual(api.addItemToListCallCount, 0)
        XCTAssertEqual(api.removeItemFromListCallCount, 1)
        XCTAssertEqual(api.listItemsCallCount, 2)
        XCTAssertNil(viewModel.items.first?.list)

        let notification = try XCTUnwrap(recorder.lastNotification)
        let userInfo = try XCTUnwrap(notification.userInfo)
        XCTAssertEqual(userInfo[AppEvents.MembershipChanged.itemIDKey] as? Int, 1)
        XCTAssertEqual(userInfo[AppEvents.MembershipChanged.isInListKey] as? Bool, false)
    }

    func test_setInListToggleOffFailure_preservesStateAndDoesNotPostNotification() async throws {
        let notificationCenter = NotificationCenter()
        let original = try decodeItem(
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
        let api = MockItemsAPI(categories: [], items: [original])
        api.removeItemFromListError = APIError.serverError("toggle off fail")

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

        let result = await viewModel.setInList(itemID: 1, isInList: false)

        XCTAssertFalse(result)
        XCTAssertEqual(api.addItemToListCallCount, 0)
        XCTAssertEqual(api.removeItemFromListCallCount, 1)
        XCTAssertEqual(api.listItemsCallCount, 1)
        XCTAssertNotNil(viewModel.items.first?.list)
        XCTAssertEqual(recorder.count, 0)
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
        api.blockedCreateCallIndices = [1]

        let parked = expectation(description: "create call parked")
        api.onCreateCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let first = Task { await viewModel.addItem(name: "Milk", categoryID: 1) }
        await fulfillment(of: [parked], timeout: 1.0)

        let second = Task { await viewModel.addItem(name: "Milk", categoryID: 1) }
        api.releaseCreateCall(1)

        _ = await first.value
        _ = await second.value

        XCTAssertEqual(api.createItemCallCount, 1)
    }

    func test_updateItem_duplicateInFlightCall_isRejected() async throws {
        let existingItem = try decodeItem(
            """
            {
              "id": 10,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )
        let updatedItem = try decodeItem(
            """
            {
              "id": 10,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Oat Milk"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [existingItem])
        api.updateItemResult = updatedItem
        api.blockedUpdateCallIndices = [1]

        let parked = expectation(description: "update call parked")
        api.onUpdateCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let first = Task { await viewModel.updateItem(id: 10, name: "Oat Milk", categoryID: 1) }
        await fulfillment(of: [parked], timeout: 1.0)
        let second = Task { await viewModel.updateItem(id: 10, name: "Oat Milk", categoryID: 1) }

        api.releaseUpdateCall(1)

        let firstResult = await first.value
        let secondResult = await second.value

        XCTAssertTrue(firstResult)
        XCTAssertFalse(secondResult)
        XCTAssertEqual(api.updateItemCallCount, 1)
    }

    func test_deleteItem_duplicateInFlightCall_isRejected() async throws {
        let item = try decodeItem(
            """
            {
              "id": 7,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Yogurt"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [item])
        api.blockedDeleteCallIndices = [1]

        let parked = expectation(description: "delete call parked")
        api.onDeleteCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let first = Task { await viewModel.deleteItem(id: 7) }
        await fulfillment(of: [parked], timeout: 1.0)
        let second = Task { await viewModel.deleteItem(id: 7) }

        api.releaseDeleteCall(1)

        let firstResult = await first.value
        let secondResult = await second.value

        XCTAssertTrue(firstResult)
        XCTAssertFalse(secondResult)
        XCTAssertEqual(api.deleteItemCallCount, 1)
    }

    func test_setInList_duplicateInFlightCall_isRejected() async throws {
        let item = try decodeItem(
            """
            {
              "id": 5,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Cream"
            }
            """
        )
        let inListItem = try decodeItem(
            """
            {
              "id": 5,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Cream",
              "list": { "id": 31, "quantity": "1", "done": false }
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [item])
        api.listItemsAfterSetInList = [inListItem]
        api.blockedAddItemToListCallIndices = [1]

        let parked = expectation(description: "addItemToList call parked")
        api.onAddItemToListCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)
        await viewModel.load()

        let first = Task { await viewModel.setInList(itemID: 5, isInList: true) }
        await fulfillment(of: [parked], timeout: 1.0)
        let second = Task { await viewModel.setInList(itemID: 5, isInList: true) }

        api.releaseAddItemToListCall(1)

        let firstResult = await first.value
        let secondResult = await second.value

        XCTAssertTrue(firstResult)
        XCTAssertFalse(secondResult)
        XCTAssertEqual(api.addItemToListCallCount, 1)
    }

    func test_setInList_rejectedWhileDeleteInFlightForSameItem() async throws {
        let item = try decodeItem(
            """
            {
              "id": 9,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Cheese"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [item])
        api.blockedDeleteCallIndices = [1]

        let parked = expectation(description: "delete call parked")
        api.onDeleteCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let deleteTask = Task { await viewModel.deleteItem(id: 9) }
        await fulfillment(of: [parked], timeout: 1.0)

        let toggleResult = await viewModel.setInList(itemID: 9, isInList: true)
        api.releaseDeleteCall(1)

        XCTAssertFalse(toggleResult)
        let deleteResult = await deleteTask.value

        XCTAssertTrue(deleteResult)
        XCTAssertEqual(api.deleteItemCallCount, 1)
        XCTAssertEqual(api.addItemToListCallCount, 0)
        XCTAssertEqual(api.removeItemFromListCallCount, 0)
    }

    func test_updateItem_rejectedWhileDeleteInFlightForSameItem() async throws {
        let item = try decodeItem(
            """
            {
              "id": 19,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Cheese"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [item])
        api.blockedDeleteCallIndices = [1]

        let parked = expectation(description: "delete call parked")
        api.onDeleteCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let deleteTask = Task { await viewModel.deleteItem(id: 19) }
        await fulfillment(of: [parked], timeout: 1.0)

        let saveResult = await viewModel.updateItem(id: 19, name: "Sharp Cheese", categoryID: 1)
        api.releaseDeleteCall(1)

        XCTAssertFalse(saveResult)
        let deleteResult = await deleteTask.value

        XCTAssertTrue(deleteResult)
        XCTAssertEqual(api.deleteItemCallCount, 1)
        XCTAssertEqual(api.updateItemCallCount, 0)
    }

    func test_editorMutations_lockEachOtherWhileUpdateInFlight() async throws {
        let existingItem = try decodeItem(
            """
            {
              "id": 14,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )
        let updatedItem = try decodeItem(
            """
            {
              "id": 14,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Oat Milk"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [existingItem])
        api.updateItemResult = updatedItem
        api.blockedUpdateCallIndices = [1]

        let parked = expectation(description: "update call parked")
        api.onUpdateCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let saveTask = Task {
            await viewModel.updateItem(id: 14, name: "Oat Milk", categoryID: 1)
        }
        await fulfillment(of: [parked], timeout: 1.0)

        let toggleResult = await viewModel.setInList(itemID: 14, isInList: true)
        let deleteResult = await viewModel.deleteItem(id: 14)

        api.releaseUpdateCall(1)
        let saveResult = await saveTask.value

        XCTAssertTrue(saveResult)
        XCTAssertFalse(toggleResult)
        XCTAssertFalse(deleteResult)
        XCTAssertEqual(api.updateItemCallCount, 1)
        XCTAssertEqual(api.addItemToListCallCount, 0)
        XCTAssertEqual(api.deleteItemCallCount, 0)
    }

    func test_editorMutations_lockEachOtherWhileToggleInFlight() async throws {
        let item = try decodeItem(
            """
            {
              "id": 15,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Cream"
            }
            """
        )
        let inListItem = try decodeItem(
            """
            {
              "id": 15,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Cream",
              "list": { "id": 40, "quantity": "1", "done": false }
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [item])
        api.listItemsAfterSetInList = [inListItem]
        api.blockedAddItemToListCallIndices = [1]

        let parked = expectation(description: "toggle call parked")
        api.onAddItemToListCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)
        await viewModel.load()

        let toggleTask = Task { await viewModel.setInList(itemID: 15, isInList: true) }
        await fulfillment(of: [parked], timeout: 1.0)

        let saveResult = await viewModel.updateItem(id: 15, name: "Whipping Cream", categoryID: 1)
        let deleteResult = await viewModel.deleteItem(id: 15)

        api.releaseAddItemToListCall(1)
        let toggleResult = await toggleTask.value

        XCTAssertTrue(toggleResult)
        XCTAssertFalse(saveResult)
        XCTAssertFalse(deleteResult)
        XCTAssertEqual(api.addItemToListCallCount, 1)
        XCTAssertEqual(api.updateItemCallCount, 0)
        XCTAssertEqual(api.deleteItemCallCount, 0)
    }

    func test_deleteItem_conflict_keepsItemAndShowsError() async throws {
        let item = try decodeItem(
            """
            {
              "id": 21,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [item])
        api.deleteItemError = APIError.conflict("Item is still referenced")
        let viewModel = ItemsViewModel(api: api)
        await viewModel.load()

        let deleteResult = await viewModel.deleteItem(id: 21)

        XCTAssertFalse(deleteResult)
        XCTAssertEqual(viewModel.items.map(\.id), [21])
        XCTAssertEqual(viewModel.errorMessage, "Item is still referenced")
    }

    func test_updateItem_notFound_keepsPriorState() async throws {
        let original = try decodeItem(
            """
            {
              "id": 33,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [original])
        api.updateItemError = APIError.notFound("Item no longer exists")
        let viewModel = ItemsViewModel(api: api)
        await viewModel.load()

        let updateResult = await viewModel.updateItem(id: 33, name: "Oat Milk", categoryID: 1)

        XCTAssertFalse(updateResult)
        XCTAssertEqual(viewModel.items.map(\.name), ["Milk"])
        XCTAssertEqual(viewModel.errorMessage, "Item no longer exists")
    }

    func test_deleteItem_notFound_keepsPriorState() async throws {
        let original = try decodeItem(
            """
            {
              "id": 34,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )

        let api = MockItemsAPI(categories: [], items: [original])
        api.deleteItemError = APIError.notFound("Item no longer exists")
        let viewModel = ItemsViewModel(api: api)
        await viewModel.load()

        let deleteResult = await viewModel.deleteItem(id: 34)

        XCTAssertFalse(deleteResult)
        XCTAssertEqual(viewModel.items.map(\.name), ["Milk"])
        XCTAssertEqual(viewModel.errorMessage, "Item no longer exists")
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

    func test_addFlowControlsDisabled_whileAddRequestInFlight() async throws {
        let api = MockItemsAPI(categories: [], items: [])
        api.createItemResult = try decodeItem(
            """
            {
              "id": 44,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )
        api.blockedCreateCallIndices = [1]

        let parked = expectation(description: "create call parked")
        api.onCreateCallParked = { callIndex in
            if callIndex == 1 {
                parked.fulfill()
            }
        }

        let viewModel = ItemsViewModel(api: api)

        let addTask = Task { await viewModel.addItem(name: "Milk", categoryID: 1) }
        await fulfillment(of: [parked], timeout: 1.0)

        XCTAssertTrue(viewModel.isAdding)
        XCTAssertTrue(viewModel.isAddButtonDisabled(name: "Milk", categoryID: 1))
        XCTAssertTrue(viewModel.isAddCategoryPickerDisabled)
        XCTAssertTrue(viewModel.isAddNameFieldDisabled)

        api.releaseCreateCall(1)
        _ = await addTask.value
    }

    func test_addFlowValidation_requiresTrimmedNameAndCategory() async throws {
        let api = MockItemsAPI(categories: [], items: [])
        let viewModel = ItemsViewModel(api: api)

        XCTAssertTrue(viewModel.isAddButtonDisabled(name: "   ", categoryID: 1))
        XCTAssertTrue(viewModel.isAddButtonDisabled(name: "Milk", categoryID: nil))
        XCTAssertFalse(viewModel.isAddButtonDisabled(name: " Milk ", categoryID: 1))
    }

    func test_itemEditorControlsDisabled_whileMutationInFlight() {
        XCTAssertTrue(ItemEditorViewUX.cancelDisabled(isMutationInFlight: true))
        XCTAssertTrue(ItemEditorViewUX.saveDisabled(isMutationInFlight: true, baseSaveDisabled: false))
        XCTAssertTrue(ItemEditorViewUX.nameDisabled(isMutationInFlight: true))
        XCTAssertTrue(ItemEditorViewUX.categoryDisabled(isMutationInFlight: true))
        XCTAssertTrue(ItemEditorViewUX.membershipToggleDisabled(isMutationInFlight: true))
        XCTAssertTrue(ItemEditorViewUX.deleteDisabled(isMutationInFlight: true))
    }

    func test_itemEditorAccessibilityLabels_remainStable() {
        XCTAssertEqual(ItemEditorViewAccessibility.categoryLabel, "Edit item category")
        XCTAssertEqual(ItemEditorViewAccessibility.nameLabel, "Edit item name")
        XCTAssertEqual(ItemEditorViewAccessibility.membershipToggleLabel, "Include item in shopping list")
        XCTAssertEqual(ItemEditorViewAccessibility.cancelButtonLabel, "Cancel edit item")
        XCTAssertEqual(ItemEditorViewAccessibility.saveButtonLabel, "Save item changes")
        XCTAssertEqual(ItemEditorViewAccessibility.deleteButtonLabel, "Delete item")
        XCTAssertEqual(ItemEditorViewAccessibility.deleteConfirmButtonLabel, "Confirm delete item")
        XCTAssertEqual(ItemEditorViewAccessibility.errorLabel, "Edit item error")
    }

    func test_setInList_notFound_keepsPriorState_andSurfacesError() async throws {
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
        api.addItemToListError = APIError.notFound("Item no longer exists")

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

        let result = await viewModel.setInList(itemID: 1, isInList: true)

        XCTAssertFalse(result)
        XCTAssertEqual(viewModel.items.map(\.id), [1])
        XCTAssertEqual(viewModel.errorMessage, "Item no longer exists")
        XCTAssertEqual(recorder.count, 0)
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
    private var notifications: [Notification] = []
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

    var lastNotification: Notification? {
        lock.lock()
        defer { lock.unlock() }
        return notifications.last
    }

    func record(_ notification: Notification) {
        lock.lock()
        notifications.append(notification)
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
    var listItemsResponsesByCall: [Int: [Item]] = [:]

    var blockedListItemsCallIndices: Set<Int> = []
    var blockedCreateCallIndices: Set<Int> = []
    var blockedUpdateCallIndices: Set<Int> = []
    var blockedDeleteCallIndices: Set<Int> = []
    var blockedAddItemToListCallIndices: Set<Int> = []

    var onListItemsCallParked: ((Int) -> Void)?
    var onCreateCallParked: ((Int) -> Void)?
    var onUpdateCallParked: ((Int) -> Void)?
    var onDeleteCallParked: ((Int) -> Void)?
    var onAddItemToListCallParked: ((Int) -> Void)?

    private var listItemsContinuations: [Int: CheckedContinuation<Void, Never>] = [:]
    private var createContinuations: [Int: CheckedContinuation<Void, Never>] = [:]
    private var updateContinuations: [Int: CheckedContinuation<Void, Never>] = [:]
    private var deleteContinuations: [Int: CheckedContinuation<Void, Never>] = [:]
    private var addItemToListContinuations: [Int: CheckedContinuation<Void, Never>] = [:]

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
        let callIndex = listItemsCallCount

        if let listItemsError {
            throw listItemsError
        }

        let response: [Item]
        if let callSpecificResponse = listItemsResponsesByCall[callIndex] {
            response = callSpecificResponse
        } else if let listItemsAfterSetInList {
            response = listItemsAfterSetInList
        } else if let inList {
            response = items.filter { ($0.list != nil) == inList }
        } else {
            response = items
        }

        if blockedListItemsCallIndices.contains(callIndex) {
            await withCheckedContinuation { continuation in
                listItemsContinuations[callIndex] = continuation
                onListItemsCallParked?(callIndex)
            }
        }

        return response
    }

    func releaseListItemsCall(_ callIndex: Int) {
        listItemsContinuations.removeValue(forKey: callIndex)?.resume()
    }

    func createItem(categoryID: Int, name: String) async throws -> Item {
        createItemCallCount += 1
        let callIndex = createItemCallCount

        if blockedCreateCallIndices.contains(callIndex) {
            await withCheckedContinuation { continuation in
                createContinuations[callIndex] = continuation
                onCreateCallParked?(callIndex)
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
        let callIndex = updateItemCallCount

        if blockedUpdateCallIndices.contains(callIndex) {
            await withCheckedContinuation { continuation in
                updateContinuations[callIndex] = continuation
                onUpdateCallParked?(callIndex)
            }
        }

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
        let callIndex = deleteItemCallCount

        if blockedDeleteCallIndices.contains(callIndex) {
            await withCheckedContinuation { continuation in
                deleteContinuations[callIndex] = continuation
                onDeleteCallParked?(callIndex)
            }
        }

        if let deleteItemError {
            throw deleteItemError
        }
        items.removeAll(where: { $0.id == id })
    }

    func addItemToList(itemID: Int) async throws -> Item {
        addItemToListCallCount += 1
        let callIndex = addItemToListCallCount

        if blockedAddItemToListCallIndices.contains(callIndex) {
            await withCheckedContinuation { continuation in
                addItemToListContinuations[callIndex] = continuation
                onAddItemToListCallParked?(callIndex)
            }
        }

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

    func releaseCreateCall(_ callIndex: Int) {
        createContinuations.removeValue(forKey: callIndex)?.resume()
    }

    func releaseUpdateCall(_ callIndex: Int) {
        updateContinuations.removeValue(forKey: callIndex)?.resume()
    }

    func releaseDeleteCall(_ callIndex: Int) {
        deleteContinuations.removeValue(forKey: callIndex)?.resume()
    }

    func releaseAddItemToListCall(_ callIndex: Int) {
        addItemToListContinuations.removeValue(forKey: callIndex)?.resume()
    }
}
