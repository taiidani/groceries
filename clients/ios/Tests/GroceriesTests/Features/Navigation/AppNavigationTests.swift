import XCTest

@testable import Aisle4

final class AppNavigationTests: XCTestCase {
    func test_appTabsIncludeListItemsAndAccount() {
        XCTAssertEqual(AppTab.allCases, [.list, .items, .account])
    }

    func test_itemsMembershipDidChangeNotificationPayloadKeys() {
        XCTAssertEqual(AppEvents.MembershipChanged.name.rawValue, "itemsMembershipDidChange")
        XCTAssertEqual(AppEvents.MembershipChanged.itemIDKey, "itemID")
        XCTAssertEqual(AppEvents.MembershipChanged.isInListKey, "isInList")
        XCTAssertEqual(AppEvents.MembershipChanged.changedAtKey, "changedAt")
    }

    func test_accountDisplayUsernameText_usesUserName() {
        XCTAssertEqual(AccountDisplay.usernameText(for: "alice"), "alice")
    }

    func test_accountDisplayUsernameText_handlesMissingUser() {
        XCTAssertEqual(AccountDisplay.usernameText(for: nil), "Unknown")
    }

    func test_accountDisplayUsernameText_trimsWhitespace() {
        XCTAssertEqual(AccountDisplay.usernameText(for: "   alice   \n"), "alice")
    }

    func test_accountDisplayUsernameText_rejectsWhitespaceOnly() {
        XCTAssertEqual(AccountDisplay.usernameText(for: "   \n  \t"), "Unknown")
    }
}
