import Foundation

// MARK: - Auth endpoints

extension GroceriesAPIClient {

    /// Authenticates with the API and returns a token + expiry.
    ///
    /// On success the client automatically stores the returned token so
    /// subsequent calls are authenticated without any extra steps.
    public func login(username: String, password: String) async throws -> LoginResponse {
        let body = LoginRequest(username: username, password: password)
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
