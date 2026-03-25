import SwiftUI

// MARK: - App tabs

enum AppTab: String, CaseIterable, Hashable {
    case list
    case items
    case account
}

enum AccountDisplay {
    static func usernameText(for username: String?) -> String {
        guard let username else {
            return "Unknown"
        }
        let trimmed = username.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            return "Unknown"
        }
        return trimmed
    }
}

struct AppTabsView: View {
    @Environment(AuthViewModel.self) private var authViewModel

    @State private var selectedTab: AppTab = .list

    var body: some View {
        TabView(selection: $selectedTab) {
            ShoppingListView(apiClient: authViewModel.apiClient)
                .tabItem {
                    Label("List", systemImage: "cart")
                }
                .tag(AppTab.list)

            ItemsView(apiClient: authViewModel.apiClient)
                .tabItem {
                    Label("Items", systemImage: "square.grid.2x2")
                }
                .tag(AppTab.items)

            AccountView()
                .tabItem {
                    Label("Account", systemImage: "person")
                }
                .tag(AppTab.account)
        }
    }
}

private struct AccountView: View {
    @Environment(AuthViewModel.self) private var authViewModel

    var body: some View {
        NavigationStack {
            List {
                Section("Profile") {
                    LabeledContent("Username", value: AccountDisplay.usernameText(for: authViewModel.currentUser?.name))
                }
            }
            .navigationTitle("Account")
        }
    }
}
