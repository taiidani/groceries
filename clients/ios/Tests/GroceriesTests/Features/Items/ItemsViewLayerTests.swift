import XCTest
import GroceriesAPI

@testable import Aisle4

@MainActor
final class ItemsViewLayerTests: XCTestCase {
    func test_itemSelectionNavigation_isNotPathDriven() {
        XCTAssertFalse(ItemsViewUX.shouldUsePathDrivenNavigation())
    }

    func test_itemSelectionRoutesToEditorByItemID() throws {
        let item = try makeItem(
            """
            {
              "id": 42,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )

        let route = ItemsViewUX.editorRoute(for: item)

        XCTAssertEqual(route, .editor(itemID: 42))
    }

    func test_editorRouteResolvesToExistingItem() throws {
        let item = try makeItem(
            """
            {
              "id": 8,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Yogurt"
            }
            """
        )

        let route = ItemsViewRoute.editor(itemID: 8)

        XCTAssertEqual(ItemsViewUX.editorItem(for: route, items: [item])?.id, 8)
    }

    func test_membershipToggleAccess_isEditorOnly() {
        XCTAssertFalse(ItemMembershipToggleAccess.isAvailable(in: .listRows))
        XCTAssertFalse(ItemMembershipToggleAccess.isAvailable(in: .addItemForm))
        XCTAssertTrue(ItemMembershipToggleAccess.isAvailable(in: .editor))
    }

    func test_retryAffordance_visibleOnlyWhenLoadFailedAndListEmpty() {
        let item = try! makeItem(
            """
            {
              "id": 1,
              "category_id": 1,
              "category_name": "Dairy",
              "name": "Milk"
            }
            """
        )

        XCTAssertTrue(
            ItemsViewUX.shouldShowRetryAffordance(
                isLoading: false,
                filteredItems: [],
                loadErrorMessage: "boom",
                mutationErrorMessage: nil
            )
        )

        XCTAssertFalse(
            ItemsViewUX.shouldShowRetryAffordance(
                isLoading: false,
                filteredItems: [item],
                loadErrorMessage: "boom",
                mutationErrorMessage: nil
            )
        )
    }

    func test_retryAffordance_hiddenWhenOnlyMutationErrorExistsAndFiltersEmptyList() {
        XCTAssertFalse(
            ItemsViewUX.shouldShowRetryAffordance(
                isLoading: false,
                filteredItems: [],
                loadErrorMessage: nil,
                mutationErrorMessage: "Category is required."
            )
        )
    }

    func test_retryAffordanceAction_invokesRetryLoadPath() async {
        let recorder = RetryActionRecorder()

        await ItemsViewUX.performRetry(using: recorder.record)

        let count = await recorder.count
        XCTAssertEqual(count, 1)
    }

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

private func makeItem(_ json: String) throws -> Item {
    let data = try XCTUnwrap(json.data(using: .utf8))
    return try JSONDecoder().decode(Item.self, from: data)
}

private actor RetryActionRecorder {
    private(set) var count = 0

    func record() async {
        count += 1
    }
}
