import XCTest

@testable import Aisle4

@MainActor
final class ItemsViewLayerTests: XCTestCase {
    func test_addFlowControlsDisabled_whileAdding() {
        XCTAssertTrue(AddItemViewUX.cancelDisabled(isAdding: true))
        XCTAssertTrue(AddItemViewUX.saveDisabled(isAdding: true, baseSaveDisabled: false))
        XCTAssertTrue(AddItemViewUX.categoryDisabled(isAdding: true))
        XCTAssertTrue(AddItemViewUX.nameDisabled(isAdding: true))
    }

    func test_interactiveDismissDisabled_whileAdding() {
        XCTAssertTrue(ItemsViewUX.addSheetInteractiveDismissDisabled(isAdding: true))
    }

    func test_accessibilityLabels_remainStable() {
        XCTAssertEqual(ItemsViewAccessibility.searchFieldLabel, "Item search")
        XCTAssertEqual(ItemsViewAccessibility.inListOnlyToggleLabel, "In List only")
        XCTAssertEqual(ItemsViewAccessibility.addItemButtonLabel, "Add item")

        XCTAssertEqual(AddItemViewAccessibility.categoryLabel, "Item category")
        XCTAssertEqual(AddItemViewAccessibility.nameLabel, "Item name")
        XCTAssertEqual(AddItemViewAccessibility.cancelButtonLabel, "Cancel add item")
        XCTAssertEqual(AddItemViewAccessibility.saveButtonLabel, "Save item")
        XCTAssertEqual(AddItemViewAccessibility.errorLabel, "Add item error")
    }

    func test_itemEditorMembershipToggle_pessimisticFlow_keepsVisibleValueUntilSuccess() {
        let initial = ItemEditorViewUX.membershipToggleInitialState(isInList: false)

        XCTAssertEqual(initial.visibleValue, false)

        let afterTap = ItemEditorViewUX.membershipToggleBeginMutation(
            currentVisibleValue: initial.visibleValue,
            requestedValue: true
        )

        XCTAssertEqual(afterTap.visibleValue, false)
        XCTAssertEqual(afterTap.requestedValue, true)

        let afterSuccess = ItemEditorViewUX.membershipToggleResolveMutation(
            currentVisibleValue: afterTap.visibleValue,
            requestedValue: afterTap.requestedValue,
            success: true
        )

        XCTAssertEqual(afterSuccess, true)
    }

    func test_itemEditorMembershipToggle_notFoundFailure_keepsVisibleValue() {
        let afterTap = ItemEditorViewUX.membershipToggleBeginMutation(
            currentVisibleValue: true,
            requestedValue: false
        )

        let afterFailure = ItemEditorViewUX.membershipToggleResolveMutation(
            currentVisibleValue: afterTap.visibleValue,
            requestedValue: afterTap.requestedValue,
            success: false
        )

        XCTAssertEqual(afterTap.visibleValue, true)
        XCTAssertEqual(afterFailure, true)
    }

    func test_itemEditorMembershipToggle_syncsToExternalModelMembershipChange() {
        let synced = ItemEditorViewUX.membershipToggleSyncExternalModelChange(
            currentVisibleValue: false,
            modelIsInList: true,
            isMutationInFlight: false
        )

        XCTAssertTrue(synced)
    }
}
