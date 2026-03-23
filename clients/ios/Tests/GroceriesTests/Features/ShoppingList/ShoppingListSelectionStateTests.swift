import XCTest
@testable import GroceriesT

final class ShoppingListSelectionStateTests: XCTestCase {
    func test_emptySelectionWithAvailableStores_selectsFirstStore() {
        let selectedStoreID = StoreSelectionReconciler.reconcile(current: nil, availableStoreIDs: [3, 7, 9])

        XCTAssertEqual(selectedStoreID, 3)
    }

    func test_removedSelection_fallsBackToFirstStoreInDeterministicOrder() {
        let selectedStoreID = StoreSelectionReconciler.reconcile(current: 42, availableStoreIDs: [1, 4, 8])

        XCTAssertEqual(selectedStoreID, 1)
    }

    func test_existingSelection_doesNotAutoAdvanceToNextStore() {
        let selectedStoreID = StoreSelectionReconciler.reconcile(current: 4, availableStoreIDs: [1, 4, 8])

        XCTAssertEqual(selectedStoreID, 4)
    }
}
