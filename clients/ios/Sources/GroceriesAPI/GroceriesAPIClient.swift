import Foundation

// MARK: - Client

/// The main entry point for all API communication.
///
/// `GroceriesAPIClient` is `actor`-isolated so that its mutable token state
/// is safe to update from any async context (e.g. after login on the main
/// actor, token refresh in the background).
public actor GroceriesAPIClient {

    // MARK: Properties

    private let baseURL: URL
    private let session: URLSession
    private var token: String?

    // MARK: Init

    /// Creates a new client.
    ///
    /// - Parameters:
    ///   - baseURL: The root URL of the API server, e.g. `http://localhost:3000`.
    ///   - session: The `URLSession` to use for requests. Defaults to a session
    ///     with no caching so responses always reflect server state.
    ///   - token: An optional pre-existing Bearer token (e.g. restored from
    ///     Keychain on app launch).
    public init(
        baseURL: URL,
        session: URLSession = .init(configuration: .ephemeral),
        token: String? = nil
    ) {
        self.baseURL = baseURL
        self.session = session
        self.token = token
    }

    // MARK: Token management

    /// Sets (or clears) the Bearer token used for authenticated requests.
    public func setToken(_ token: String?) {
        self.token = token
    }

    /// Returns `true` when the client has a token set.
    public var isAuthenticated: Bool {
        token != nil
    }

    // MARK: - Request builders

    /// Builds a `URLRequest` for the given path and HTTP method, attaching
    /// the Bearer token and JSON content-type where appropriate.
    func request(
        method: String,
        path: String,
        queryItems: [URLQueryItem]? = nil
    ) throws -> URLRequest {
        guard
            var components = URLComponents(
                url: baseURL.appendingPathComponent(path),
                resolvingAgainstBaseURL: false
            )
        else {
            throw APIError.badRequest("Could not construct URL for path: \(path)")
        }

        if let queryItems, !queryItems.isEmpty {
            components.queryItems = queryItems
        }

        guard let url = components.url else {
            throw APIError.badRequest("Could not resolve URL for path: \(path)")
        }

        var req = URLRequest(url: url)
        req.httpMethod = method

        if let token {
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        return req
    }

    /// Builds a `URLRequest` with a JSON-encoded body.
    func request<Body: Encodable>(
        method: String,
        path: String,
        body: Body,
        encoder: JSONEncoder = .apiEncoder
    ) throws -> URLRequest {
        var req = try request(method: method, path: path)
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try encoder.encode(body)
        return req
    }

    // MARK: - Response handling

    /// Executes a request and decodes a `Decodable` response body.
    func perform<Response: Decodable>(
        _ urlRequest: URLRequest,
        decoder: JSONDecoder = .apiDecoder
    ) async throws -> Response {
        let (data, response) = try await session.data(for: urlRequest)
        #if DEBUG
            let rawBody = String(data: data, encoding: .utf8) ?? "<non-UTF8 body>"
            let statusCode = (response as? HTTPURLResponse)?.statusCode ?? -1
            print(
                "[GroceriesAPI] \(urlRequest.httpMethod ?? "?") \(urlRequest.url?.path ?? "?") → HTTP \(statusCode)"
            )
            print("[GroceriesAPI] Response body: \(rawBody)")
        #endif
        try validateResponse(response, data: data)
        do {
            return try decoder.decode(Response.self, from: data)
        } catch {
            #if DEBUG
                print("[GroceriesAPI] Decode error for \(Response.self): \(error)")
            #endif
            throw APIError.decodingError(error)
        }
    }

    /// Executes a request that returns no body (e.g. 204 No Content).
    func performVoid(_ urlRequest: URLRequest) async throws {
        let (data, response) = try await session.data(for: urlRequest)
        try validateResponse(response, data: data)
    }

    /// Maps HTTP error status codes to typed `APIError` values.
    private func validateResponse(_ response: URLResponse, data: Data) throws {
        guard let http = response as? HTTPURLResponse else { return }

        guard http.statusCode < 200 || http.statusCode >= 300 else { return }

        // Attempt to read a structured error message from the body.
        let message: String
        if let apiErr = try? JSONDecoder().decode(APIErrorResponse.self, from: data) {
            message = apiErr.error
        } else {
            message = HTTPURLResponse.localizedString(forStatusCode: http.statusCode)
        }

        switch http.statusCode {
        case 400:
            throw APIError.badRequest(message)
        case 401:
            throw APIError.unauthorized
        case 403:
            throw APIError.forbidden
        case 404:
            throw APIError.notFound(message)
        case 409:
            throw APIError.conflict(message)
        case 500...:
            throw APIError.serverError(message)
        default:
            throw APIError.unexpectedStatus(http.statusCode)
        }
    }
}

// MARK: - JSONEncoder / JSONDecoder helpers

extension JSONEncoder {
    /// A shared encoder configured to match the API's snake_case conventions.
    static let apiEncoder: JSONEncoder = {
        let enc = JSONEncoder()
        enc.outputFormatting = .sortedKeys
        return enc
    }()
}

extension JSONDecoder {
    /// A shared decoder configured to handle the API's snake_case keys and
    /// RFC 3339 date strings as emitted by Go's `encoding/json`, which
    /// includes fractional seconds and a timezone offset
    /// (e.g. `"2026-04-06T22:13:09.158845-07:00"`).
    ///
    /// Swift's built-in `.iso8601` strategy cannot handle fractional seconds,
    /// so we use a custom `ISO8601DateFormatter` with `.withFractionalSeconds`.
    static let apiDecoder: JSONDecoder = {
        let dec = JSONDecoder()

        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [
            .withInternetDateTime,
            .withFractionalSeconds,
        ]

        dec.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let string = try container.decode(String.self)

            if let date = formatter.date(from: string) {
                return date
            }

            // Fall back to plain ISO 8601 without fractional seconds in case
            // the server ever emits a date without them (e.g. on the second boundary).
            let fallback = ISO8601DateFormatter()
            fallback.formatOptions = [.withInternetDateTime]
            if let date = fallback.date(from: string) {
                return date
            }

            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Cannot decode date from string: \(string)"
            )
        }

        return dec
    }()
}
