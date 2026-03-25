import SwiftUI

// MARK: - GroceriesApp

@main
struct GroceriesApp: App {

    // MARK: - Environment

    @State private var authViewModel: AuthViewModel

    // MARK: - Init

    init() {
        // Resolve the API base URL from the bundle's Info.plist so it can be
        // overridden per-scheme in Xcode without recompiling.
        // Falls back to localhost for local development.
        let baseURLString =
            Bundle.main.object(forInfoDictionaryKey: "API_BASE_URL") as? String
            ?? "http://localhost:3000"

        let baseURL = URL(string: baseURLString) ?? URL(string: "http://localhost:3000")!
        _authViewModel = State(initialValue: AuthViewModel(baseURL: baseURL))
    }

    // MARK: - Scene

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(authViewModel)
                .task {
                    // On launch, silently validate any restored Keychain token.
                    // If it has expired the user will be redirected to LoginView.
                    await authViewModel.refreshCurrentUser()
                }
        }
    }
}

// MARK: - RootView

/// The top-level routing view.
///
/// Switches between `LoginView` and `ShoppingListView` based on authentication
/// state. Uses `@Environment` rather than prop drilling so any descendant can
/// observe auth state without additional wiring.
struct RootView: View {

    @Environment(AuthViewModel.self) private var authViewModel

    var body: some View {
        if authViewModel.isAuthenticated {
            AppTabsView()
        } else {
            LoginView()
        }
    }
}
