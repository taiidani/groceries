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

    func test_refreshRequestGate_notificationBurst_collapsesToOneRefreshRequest() {
        var gate = ShoppingListRefreshRequestGate()
        let now = Date()

        let first = gate.shouldStartRefresh(trigger: .membershipChanged, now: now)
        gate.refreshCompleted()
        let second = gate.shouldStartRefresh(
            trigger: .membershipChanged,
            now: now.addingTimeInterval(0.1)
        )
        let third = gate.shouldStartRefresh(
            trigger: .membershipChanged,
            now: now.addingTimeInterval(0.2)
        )

        XCTAssertTrue(first)
        XCTAssertFalse(second)
        XCTAssertFalse(third)
    }

    func test_refreshRequestGate_notificationDuringInFlight_doesNotOverlapRefresh() {
        var gate = ShoppingListRefreshRequestGate()
        let now = Date()

        let first = gate.shouldStartRefresh(trigger: .membershipChanged, now: now)
        let second = gate.shouldStartRefresh(
            trigger: .membershipChanged,
            now: now.addingTimeInterval(0.4)
        )

        XCTAssertTrue(first)
        XCTAssertFalse(second)
    }

    func test_refreshRequestGate_onAppearAndNotification_shareGateLogic() {
        var gate = ShoppingListRefreshRequestGate()
        let now = Date()

        let onAppear = gate.shouldStartRefresh(trigger: .onAppear, now: now)
        gate.refreshCompleted()
        let notification = gate.shouldStartRefresh(
            trigger: .membershipChanged,
            now: now.addingTimeInterval(0.1)
        )

        XCTAssertTrue(onAppear)
        XCTAssertFalse(notification)
    }

    func test_refreshRequestGate_notificationAndImmediateTabSwitchRace_triggersExactlyOneRefresh() {
        var gate = ShoppingListRefreshRequestGate()
        let now = Date()

        let notification = gate.shouldStartRefresh(trigger: .membershipChanged, now: now)
        let tabSwitchAppear = gate.shouldStartRefresh(
            trigger: .onAppear,
            now: now.addingTimeInterval(0.01)
        )

        XCTAssertTrue(notification)
        XCTAssertFalse(tabSwitchAppear)
    }

    func test_membershipObserver_repeatedAppearDisappearCycles_doNotRegisterDuplicates() {
        let center = NotificationCenter()
        let probe = MembershipRefreshObserverProbe(notificationCenter: center)
        let counter = CallbackCounter()

        for _ in 0..<3 {
            probe.start {
                counter.increment()
            }
            center.post(name: AppEvents.MembershipChanged.name, object: nil)
            probe.stop()
        }

        XCTAssertEqual(counter.value, 3)
    }
}

private final class CallbackCounter: @unchecked Sendable {
    private let lock = NSLock()
    private var storage = 0

    var value: Int {
        lock.lock()
        defer { lock.unlock() }
        return storage
    }

    func increment() {
        lock.lock()
        storage += 1
        lock.unlock()
    }
}

private final class MembershipRefreshObserverProbe {
    private let observer: ShoppingListMembershipRefreshObserver

    init(notificationCenter: NotificationCenter) {
        observer = ShoppingListMembershipRefreshObserver(notificationCenter: notificationCenter)
    }

    func start(onMembershipChanged: @escaping @Sendable () -> Void) {
        observer.start(onMembershipChanged: onMembershipChanged)
    }

    func stop() {
        observer.stop()
    }
}
