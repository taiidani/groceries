import XCTest
@testable import GroceriesT

final class StoreSelectionReconcilerTests: XCTestCase {
    func test_keepsSelectionWhenStillAvailable() {
        let selected = StoreSelectionReconciler.reconcile(current: 2, availableStoreIDs: [1, 2, 3])
        XCTAssertEqual(selected, 2)
    }

    func test_fallsBackToFirstAvailableWhenSelectionMissing() {
        let selected = StoreSelectionReconciler.reconcile(current: 4, availableStoreIDs: [1, 2, 3])
        XCTAssertEqual(selected, 1)
    }

    func test_returnsNilWhenNoStoresAvailable() {
        let selected = StoreSelectionReconciler.reconcile(current: 2, availableStoreIDs: [])
        XCTAssertNil(selected)
    }

    func test_selectsFirstWhenCurrentIsNilAndStoresExist() {
        let selected = StoreSelectionReconciler.reconcile(current: nil, availableStoreIDs: [9, 11])
        XCTAssertEqual(selected, 9)
    }

    func test_keepsCurrentEvenWhenNotFirst() {
        let selected = StoreSelectionReconciler.reconcile(current: 11, availableStoreIDs: [9, 11, 12])
        XCTAssertEqual(selected, 11)
    }
}
