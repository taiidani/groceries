import XCTest
@testable import GroceriesT

final class AppNavigationTests: XCTestCase {
    func test_appTabsIncludeListAndAccount() {
        XCTAssertEqual(AppTab.allCases, [.list, .account])
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
