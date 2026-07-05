import GroceriesAPI
import XCTest

@testable import Aisle4

final class ShoppingListEditorNavigationTests: XCTestCase {
    func test_isLongPressEditAvailable_falseWhileMutating() {
        XCTAssertFalse(ShoppingListEditorUX.isLongPressEditAvailable(isMutating: true))
        XCTAssertTrue(ShoppingListEditorUX.isLongPressEditAvailable(isMutating: false))
    }

    func test_editorItem_resolvesMatchingItemByID() {
        let items = [
            Item(id: 1, categoryID: 10, categoryName: "Dairy", name: "Milk", list: nil),
            Item(id: 2, categoryID: 20, categoryName: "Bakery", name: "Bread", list: nil),
        ]

        let resolved = ShoppingListEditorUX.editorItem(for: 2, in: items)

        XCTAssertEqual(resolved?.id, 2)
        XCTAssertEqual(resolved?.name, "Bread")
    }

    func test_editorItem_returnsNilWhenNoMatchExists() {
        let items = [
            Item(id: 1, categoryID: 10, categoryName: "Dairy", name: "Milk", list: nil)
        ]

        XCTAssertNil(ShoppingListEditorUX.editorItem(for: 999, in: items))
    }

    func test_longPressMinimumDuration_isPositiveAndBriefEnoughToFeelResponsive() {
        XCTAssertGreaterThan(ShoppingListEditorUX.longPressMinimumDuration, 0)
        XCTAssertLessThanOrEqual(ShoppingListEditorUX.longPressMinimumDuration, 1.0)
    }

    func test_accessibilityLabels_remainStable() {
        XCTAssertEqual(ShoppingListRowAccessibility.editActionLabel, "Edit item")
    }
}
