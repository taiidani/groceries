import XCTest

@testable import Aisle4

final class ShoppingListAutoRefreshCoordinatorTests: XCTestCase {
    func test_appDidBecomeActive_whenIdle_refreshesImmediately() {
        var coordinator = ShoppingListAutoRefreshCoordinator()

        let action = coordinator.appDidBecomeActive(isBusy: false)

        XCTAssertEqual(action, .refreshNow)
        XCTAssertFalse(coordinator.pendingRefresh)
    }

    func test_appDidBecomeActive_whenBusy_defersUntilIdle() {
        var coordinator = ShoppingListAutoRefreshCoordinator()

        let activeAction = coordinator.appDidBecomeActive(isBusy: true)
        let stillBusyAction = coordinator.mutationStateDidChange(isBusy: true)
        let idleAction = coordinator.mutationStateDidChange(isBusy: false)

        XCTAssertEqual(activeAction, .deferUntilIdle)
        XCTAssertEqual(stillBusyAction, .none)
        XCTAssertEqual(idleAction, .refreshNow)
        XCTAssertFalse(coordinator.pendingRefresh)
    }

    func test_pendingRefresh_isConsumedByIdleActivation() {
        var coordinator = ShoppingListAutoRefreshCoordinator()

        _ = coordinator.appDidBecomeActive(isBusy: true)
        let action = coordinator.appDidBecomeActive(isBusy: false)

        XCTAssertEqual(action, .refreshNow)
        XCTAssertFalse(coordinator.pendingRefresh)
    }
}
