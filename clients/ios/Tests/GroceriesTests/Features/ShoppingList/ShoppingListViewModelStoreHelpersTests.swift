import XCTest
@testable import GroceriesT
import GroceriesAPI

@MainActor
final class ShoppingListViewModelStoreHelpersTests: XCTestCase {
    func test_nonEmptyStoreGroups_excludesEmptyStores() throws {
        let viewModel = makeViewModel()

        viewModel._injectTestData(
            stores: try decodeStores(
                """
                [
                  { "id": 1, "name": "Produce" },
                  { "id": 2, "name": "Bakery" }
                ]
                """
            ),
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 10,
                    "store_id": 1,
                    "name": "Fruit",
                    "description": "",
                    "item_count": 0
                  },
                  {
                    "id": 20,
                    "store_id": 2,
                    "name": "Bread",
                    "description": "",
                    "item_count": 0
                  }
                ]
                """
            ),
            items: [
                ListItem(id: 100, itemID: 1000, itemName: "Apple", categoryID: 10, quantity: "1", done: false)
            ]
        )

        XCTAssertEqual(viewModel.nonEmptyStoreGroups.map(\.id), [1])
    }

    func test_isStoreComplete_trueOnlyWhenAllDone() throws {
        let viewModel = makeViewModel()

        viewModel._injectTestData(
            stores: try decodeStores(
                """
                [
                  { "id": 1, "name": "Produce" },
                  { "id": 2, "name": "Bakery" }
                ]
                """
            ),
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 10,
                    "store_id": 1,
                    "name": "Fruit",
                    "description": "",
                    "item_count": 0
                  },
                  {
                    "id": 20,
                    "store_id": 2,
                    "name": "Bread",
                    "description": "",
                    "item_count": 0
                  }
                ]
                """
            ),
            items: [
                ListItem(id: 100, itemID: 1000, itemName: "Apple", categoryID: 10, quantity: "1", done: true),
                ListItem(id: 101, itemID: 1001, itemName: "Pear", categoryID: 10, quantity: "1", done: true),
                ListItem(id: 200, itemID: 2000, itemName: "Sourdough", categoryID: 20, quantity: "1", done: false)
            ]
        )

        XCTAssertTrue(viewModel.isStoreComplete(storeID: 1))
        XCTAssertFalse(viewModel.isStoreComplete(storeID: 2))
        XCTAssertFalse(viewModel.isStoreComplete(storeID: 999))
    }

    func test_storeTotals_countsTotalAndDoneItemsPerStore() throws {
        let viewModel = makeViewModel()

        viewModel._injectTestData(
            stores: try decodeStores(
                """
                [
                  { "id": 1, "name": "Produce" },
                  { "id": 2, "name": "Bakery" }
                ]
                """
            ),
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 10,
                    "store_id": 1,
                    "name": "Fruit",
                    "description": "",
                    "item_count": 0
                  },
                  {
                    "id": 20,
                    "store_id": 2,
                    "name": "Bread",
                    "description": "",
                    "item_count": 0
                  }
                ]
                """
            ),
            items: [
                ListItem(id: 100, itemID: 1000, itemName: "Apple", categoryID: 10, quantity: "1", done: true),
                ListItem(id: 101, itemID: 1001, itemName: "Pear", categoryID: 10, quantity: "1", done: false),
                ListItem(id: 200, itemID: 2000, itemName: "Sourdough", categoryID: 20, quantity: "1", done: false)
            ]
        )

        let produceTotals = viewModel.storeTotals(storeID: 1)
        XCTAssertEqual(produceTotals.total, 2)
        XCTAssertEqual(produceTotals.done, 1)

        let bakeryTotals = viewModel.storeTotals(storeID: 2)
        XCTAssertEqual(bakeryTotals.total, 1)
        XCTAssertEqual(bakeryTotals.done, 0)

        let unknownTotals = viewModel.storeTotals(storeID: 999)
        XCTAssertEqual(unknownTotals.total, 0)
        XCTAssertEqual(unknownTotals.done, 0)
    }

    func test_nonEmptyStoreGroups_hasDeterministicOrdering() throws {
        let viewModel = makeViewModel()

        viewModel._injectTestData(
            stores: try decodeStores(
                """
                [
                  { "id": 2, "name": "Zeta" },
                  { "id": 1, "name": "Alpha" }
                ]
                """
            ),
            categories: try decodeCategories(
                """
                [
                  {
                    "id": 20,
                    "store_id": 2,
                    "name": "Zed Cat",
                    "description": "",
                    "item_count": 0
                  },
                  {
                    "id": 10,
                    "store_id": 1,
                    "name": "A Cat",
                    "description": "",
                    "item_count": 0
                  }
                ]
                """
            ),
            items: [
                ListItem(id: 100, itemID: 1000, itemName: "Apple", categoryID: 10, quantity: "1", done: false),
                ListItem(id: 200, itemID: 2000, itemName: "Bread", categoryID: 20, quantity: "1", done: false)
            ]
        )

        XCTAssertEqual(viewModel.nonEmptyStoreGroups.map(\.id), [1, 2])
        XCTAssertEqual(viewModel.nonEmptyStoreGroups.map(\.name), ["Alpha", "Zeta"])
    }

    private func makeViewModel() -> ShoppingListViewModel {
        ShoppingListViewModel(apiClient: GroceriesAPIClient(baseURL: URL(string: "http://localhost:3000")!))
    }

    private func decodeStores(_ json: String) throws -> [Store] {
        let data = try XCTUnwrap(json.data(using: .utf8))
        return try JSONDecoder().decode([Store].self, from: data)
    }

    private func decodeCategories(_ json: String) throws -> [GroceriesAPI.Category] {
        let data = try XCTUnwrap(json.data(using: .utf8))
        return try JSONDecoder().decode([GroceriesAPI.Category].self, from: data)
    }
}
