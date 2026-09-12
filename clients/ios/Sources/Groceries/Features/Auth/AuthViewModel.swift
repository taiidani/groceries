import AuthenticationServices
import Foundation
import GroceriesAPI

// MARK: - AuthViewModel

/// Owns the shared `GroceriesAPIClient` instance and manages authentication
/// state for the entire app.
///
/// Marked `@Observable` so SwiftUI views automatically re-render when
/// `isAuthenticated` or `currentUser` changes. Stored as an environment
/// object at the root of the view hierarchy.
@Observable
@MainActor
final class AuthViewModel {

    // MARK: - State

    /// Whether the user is currently authenticated.
    private(set) var isAuthenticated: Bool = false

    /// The currently signed-in user, populated after a successful login or
    /// on app launch when a stored token is restored from the Keychain.
    private(set) var currentUser: User?

    /// Non-nil while an async operation is in flight.
    private(set) var isLoading: Bool = false

    /// The most recent error to display to the user, cleared on the next
    /// successful operation.
    private(set) var errorMessage: String?

    // MARK: - Dependencies

    /// The shared API client. Exposed so child view models can use it
    /// without receiving it through multiple layers of initialiser injection.
    let apiClient: GroceriesAPIClient

    /// Drives the native OIDC login flow against Authelia.
    private let oidcAuthenticator = OIDCAuthenticator()

    // MARK: - Init

    /// Creates an `AuthViewModel`.
    ///
    /// On initialisation the Keychain is consulted for a previously stored
    /// token; if one is found the client is pre-authenticated and
    /// `isAuthenticated` is set to `true` optimistically. Call
    /// `refreshCurrentUser()` afterwards to confirm the token is still valid.
    ///
    /// - Parameter baseURL: The root URL of the API server.
    init(baseURL: URL) {
        let storedToken = try? KeychainStore.loadToken()
        self.apiClient = GroceriesAPIClient(baseURL: baseURL, token: storedToken)
        self.isAuthenticated = storedToken != nil
    }

    // MARK: - Public API

    /// Signs the user in via Authelia: presents the system web view for the
    /// user to authenticate, exchanges the resulting Authelia access token
    /// for this app's own Bearer token, then fetches the current user.
    func login() async {
        guard !isLoading else { return }

        isLoading = true
        errorMessage = nil

        defer { isLoading = false }

        do {
            let accessToken = try await oidcAuthenticator.authenticate()
            let response = try await apiClient.login(accessToken: accessToken)

            // Persist token to Keychain.
            try KeychainStore.saveToken(response.token)

            // Fetch the current user to populate the profile.
            let user = try await apiClient.me()

            // Commit all state changes together.
            currentUser = user
            isAuthenticated = true
        } catch let apiError as APIError {
            errorMessage = apiError.errorDescription
        } catch is CancellationError {
            // User dismissed the web authentication sheet — not an error.
        } catch let authError as ASWebAuthenticationSessionError where authError.code == .canceledLogin {
            // User dismissed the web authentication sheet — not an error.
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    /// Signs the current user out: invalidates the server-side token,
    /// removes it from the Keychain, and resets local state.
    func logout() async {
        guard !isLoading else { return }

        isLoading = true
        errorMessage = nil

        defer { isLoading = false }

        do {
            // Best-effort server-side logout; don't block local cleanup on
            // failure (e.g. if the token is already expired).
            try await apiClient.logout()
        } catch {
            // Log but continue — local state must still be cleared.
            #if DEBUG
                print("[AuthViewModel] Server logout error (ignored): \(error)")
            #endif
        }

        do {
            try KeychainStore.deleteToken()
        } catch {
            #if DEBUG
                print("[AuthViewModel] Keychain delete error: \(error)")
            #endif
        }

        currentUser = nil
        isAuthenticated = false
    }

    /// Re-validates the stored token by fetching the current user from the
    /// API. Call this on app launch when `isAuthenticated` is `true` due to
    /// a restored Keychain token.
    ///
    /// If the server rejects the token (e.g. it expired while the app was
    /// in the background) `isAuthenticated` is set to `false` and the stale
    /// token is removed from the Keychain.
    func refreshCurrentUser() async {
        guard isAuthenticated else { return }

        do {
            let user = try await apiClient.me()
            currentUser = user
        } catch APIError.unauthorized {
            // Token is no longer valid — force a fresh login.
            try? KeychainStore.deleteToken()
            await apiClient.setToken(nil)
            currentUser = nil
            isAuthenticated = false
        } catch {
            // Network error etc. — leave authenticated state as-is so the
            // user isn't unexpectedly signed out on a flaky connection.
            #if DEBUG
                print("[AuthViewModel] refreshCurrentUser error: \(error)")
            #endif
        }
    }

    /// Clears the currently displayed error message.
    func clearError() {
        errorMessage = nil
    }
}
