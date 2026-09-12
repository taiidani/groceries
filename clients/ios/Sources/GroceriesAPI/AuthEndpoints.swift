import Foundation

// MARK: - Auth endpoints

extension GroceriesAPIClient {

    /// Authenticates with the API by exchanging an Authelia access token,
    /// returning the app's own token + expiry.
    ///
    /// On success the client automatically stores the returned token so
    /// subsequent calls are authenticated without any extra steps.
    ///
    /// - Parameter accessToken: An OAuth 2.0 access token issued by Authelia
    ///   after the app completes its own OIDC login (Authorization Code +
    ///   PKCE via `ASWebAuthenticationSession`).
    public func login(accessToken: String) async throws -> LoginResponse {
        let body = AccessTokenLoginRequest(accessToken: accessToken)
        let req = try request(method: "POST", path: "/api/v1/auth/login", body: body)
        let response: LoginResponse = try await perform(req)
        setToken(response.token)
        return response
    }

    /// Invalidates the current token on the server and clears it locally.
    public func logout() async throws {
        let req = try request(method: "POST", path: "/api/v1/auth/logout")
        try await performVoid(req)
        setToken(nil)
    }

    /// Returns the currently authenticated user.
    public func me() async throws -> User {
        let req = try request(method: "GET", path: "/api/v1/auth/me")
        return try await perform(req)
    }
}
