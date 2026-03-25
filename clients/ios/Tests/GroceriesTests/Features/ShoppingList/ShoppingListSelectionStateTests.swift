import XCTest

@testable import Aisle4

final class ShoppingListSelectionStateTests: XCTestCase {
    func test_selectionTransitionsAcrossStoreSnapshots() {
        var current: Int?
        var observed: [Int?] = []

        current = StoreSelectionReconciler.reconcile(current: current, availableStoreIDs: [])
        observed.append(current)

        current = StoreSelectionReconciler.reconcile(current: current, availableStoreIDs: [5, 9])
        observed.append(current)

        current = StoreSelectionReconciler.reconcile(current: current, availableStoreIDs: [5, 9])
        observed.append(current)

        current = StoreSelectionReconciler.reconcile(current: current, availableStoreIDs: [9])
        observed.append(current)

        current = StoreSelectionReconciler.reconcile(current: current, availableStoreIDs: [])
        observed.append(current)

        XCTAssertEqual(observed, [nil, 5, 5, 9, nil])
    }

    func test_selectionDoesNotAutoAdvanceWhenCurrentRemainsPresentAcrossReordering() {
        var current: Int? = nil
        var observed: [Int?] = []

        current = StoreSelectionReconciler.reconcile(
            current: current, availableStoreIDs: [2, 8, 10])
        observed.append(current)

        current = StoreSelectionReconciler.reconcile(
            current: current, availableStoreIDs: [10, 2, 8])
        observed.append(current)

        current = StoreSelectionReconciler.reconcile(
            current: current, availableStoreIDs: [8, 2, 10])
        observed.append(current)

        XCTAssertEqual(observed, [2, 2, 2])
    }
}
