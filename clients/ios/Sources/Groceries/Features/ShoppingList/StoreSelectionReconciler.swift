enum StoreSelectionReconciler {
    static func reconcile(current: Int?, availableStoreIDs: [Int]) -> Int? {
        guard !availableStoreIDs.isEmpty else { return nil }

        if let current, availableStoreIDs.contains(current) {
            return current
        }

        return availableStoreIDs.first
    }
}
