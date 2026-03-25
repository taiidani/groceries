import Foundation

enum AppEvents {
    enum MembershipChanged {
        static let name = Notification.Name("itemsMembershipDidChange")
        static let itemIDKey = "itemID"
        static let isInListKey = "isInList"
        static let changedAtKey = "changedAt"
    }
}
