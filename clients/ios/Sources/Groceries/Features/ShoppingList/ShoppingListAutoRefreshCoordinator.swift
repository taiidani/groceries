import Foundation

struct ShoppingListAutoRefreshCoordinator {
    enum Action: Equatable {
        case none
        case deferUntilIdle
        case refreshNow
    }

    private(set) var pendingRefresh = false

    mutating func appDidBecomeActive(isBusy: Bool) -> Action {
        if isBusy {
            pendingRefresh = true
            return .deferUntilIdle
        }

        pendingRefresh = false
        return .refreshNow
    }

    mutating func mutationStateDidChange(isBusy: Bool) -> Action {
        guard pendingRefresh, !isBusy else {
            return .none
        }

        pendingRefresh = false
        return .refreshNow
    }
}
