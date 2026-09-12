import AuthenticationServices
import CryptoKit
import Foundation
import UIKit

// MARK: - OIDCAuthenticator

/// Drives the native OIDC login flow against Authelia: Authorization Code +
/// PKCE via `ASWebAuthenticationSession`, following RFC 8252 (OAuth 2.0 for
/// Native Apps).
///
/// Unlike the Obsidian plugin (which uses the Device Authorization Grant
/// because it has no redirect target), the iOS app can present a system web
/// view and receive a redirect via a custom URL scheme, so it uses the same
/// Authorization Code + PKCE flow as the web app.
@MainActor
final class OIDCAuthenticator: NSObject {

    // MARK: - Configuration

    /// Authelia's issuer / base URL. There is only one Authelia instance, so
    /// this is hardcoded rather than made configurable (mirrors the
    /// Obsidian plugin's approach in `settings.js`).
    private static let issuerURL = "https://auth.taiidani.com"
    private static let clientID = "groceries-ios"
    private static let redirectURI = "com.ryannixon.groceries://auth/callback"
    private static let callbackURLScheme = "com.ryannixon.groceries"
    private static let scope = "openid profile email groups"

    // MARK: - Errors

    enum OIDCError: Error, LocalizedError {
        case invalidAuthorizationURL
        case invalidRedirect
        case missingCode
        case stateMismatch
        case tokenExchangeFailed(String)
        case missingAccessToken

        var errorDescription: String? {
            switch self {
            case .invalidAuthorizationURL:
                return "Could not construct the Authelia authorization URL."
            case .invalidRedirect:
                return "Authelia did not return a valid redirect."
            case .missingCode:
                return "Authelia's response did not include an authorization code."
            case .stateMismatch:
                return "The authentication response could not be verified. Please try again."
            case .tokenExchangeFailed(let message):
                return "Could not complete sign-in: \(message)"
            case .missingAccessToken:
                return "Authelia's token response did not include an access token."
            }
        }
    }

    // MARK: - Public API

    /// Runs the full Authorization Code + PKCE flow and returns the Authelia
    /// access token on success.
    func authenticate() async throws -> String {
        let verifier = Self.randomURLSafeString(length: 64)
        let challenge = Self.codeChallenge(forVerifier: verifier)
        let state = Self.randomURLSafeString(length: 32)

        let redirectURL = try await presentAuthorizationSession(
            state: state,
            codeChallenge: challenge
        )

        let (code, returnedState) = try Self.extractCodeAndState(from: redirectURL)
        guard returnedState == state else {
            throw OIDCError.stateMismatch
        }

        return try await exchangeCodeForAccessToken(code: code, verifier: verifier)
    }

    // MARK: - Authorization step

    private func presentAuthorizationSession(state: String, codeChallenge: String) async throws -> URL {
        guard var components = URLComponents(string: "\(Self.issuerURL)/api/oidc/authorization") else {
            throw OIDCError.invalidAuthorizationURL
        }

        components.queryItems = [
            URLQueryItem(name: "client_id", value: Self.clientID),
            URLQueryItem(name: "response_type", value: "code"),
            URLQueryItem(name: "redirect_uri", value: Self.redirectURI),
            URLQueryItem(name: "scope", value: Self.scope),
            URLQueryItem(name: "state", value: state),
            URLQueryItem(name: "code_challenge", value: codeChallenge),
            URLQueryItem(name: "code_challenge_method", value: "S256"),
        ]

        guard let url = components.url else {
            throw OIDCError.invalidAuthorizationURL
        }

        return try await withCheckedThrowingContinuation { continuation in
            let session = ASWebAuthenticationSession(
                url: url,
                callbackURLScheme: Self.callbackURLScheme
            ) { callbackURL, error in
                if let error {
                    continuation.resume(throwing: error)
                    return
                }
                guard let callbackURL else {
                    continuation.resume(throwing: OIDCError.invalidRedirect)
                    return
                }
                continuation.resume(returning: callbackURL)
            }

            session.presentationContextProvider = self
            session.prefersEphemeralWebBrowserSession = false
            if !session.start() {
                continuation.resume(throwing: OIDCError.invalidRedirect)
            }
        }
    }

    // MARK: - Token exchange

    private func exchangeCodeForAccessToken(code: String, verifier: String) async throws -> String {
        guard let tokenURL = URL(string: "\(Self.issuerURL)/api/oidc/token") else {
            throw OIDCError.invalidAuthorizationURL
        }

        var request = URLRequest(url: tokenURL)
        request.httpMethod = "POST"
        request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")

        let bodyParams = [
            "grant_type": "authorization_code",
            "client_id": Self.clientID,
            "code": code,
            "redirect_uri": Self.redirectURI,
            "code_verifier": verifier,
        ]
        request.httpBody = Self.formEncode(bodyParams).data(using: .utf8)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            let message = String(data: data, encoding: .utf8) ?? "unknown error"
            throw OIDCError.tokenExchangeFailed(message)
        }

        struct TokenResponse: Decodable {
            let accessToken: String

            enum CodingKeys: String, CodingKey {
                case accessToken = "access_token"
            }
        }

        guard let decoded = try? JSONDecoder().decode(TokenResponse.self, from: data) else {
            throw OIDCError.missingAccessToken
        }

        return decoded.accessToken
    }

    // MARK: - Helpers

    private static func extractCodeAndState(from url: URL) throws -> (code: String, state: String) {
        guard
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
            let queryItems = components.queryItems
        else {
            throw OIDCError.invalidRedirect
        }

        guard let code = queryItems.first(where: { $0.name == "code" })?.value else {
            throw OIDCError.missingCode
        }

        let state = queryItems.first(where: { $0.name == "state" })?.value ?? ""
        return (code, state)
    }

    private static func randomURLSafeString(length: Int) -> String {
        var bytes = [UInt8](repeating: 0, count: length)
        _ = SecRandomCopyBytes(kSecRandomDefault, length, &bytes)
        return base64URLEncode(Data(bytes))
    }

    private static func codeChallenge(forVerifier verifier: String) -> String {
        let digest = SHA256.hash(data: Data(verifier.utf8))
        return base64URLEncode(Data(digest))
    }

    private static func base64URLEncode(_ data: Data) -> String {
        data.base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }

    private static func formEncode(_ params: [String: String]) -> String {
        params.map { key, value in
            let encodedKey = key.addingPercentEncoding(withAllowedCharacters: .urlQueryValueAllowed) ?? key
            let encodedValue = value.addingPercentEncoding(withAllowedCharacters: .urlQueryValueAllowed) ?? value
            return "\(encodedKey)=\(encodedValue)"
        }.joined(separator: "&")
    }
}

// MARK: - ASWebAuthenticationPresentationContextProviding

extension OIDCAuthenticator: ASWebAuthenticationPresentationContextProviding {
    func presentationAnchor(for session: ASWebAuthenticationSession) -> ASPresentationAnchor {
        let scenes = UIApplication.shared.connectedScenes
        let windowScene =
            (scenes.first(where: { $0.activationState == .foregroundActive }) as? UIWindowScene)
            ?? scenes.first(where: { $0 is UIWindowScene }) as? UIWindowScene

        guard let windowScene else {
            // No window scene is available at all (should not happen while presenting
            // a web authentication session). ASPresentationAnchor's parameterless
            // initializer is deprecated, so there is no safe fallback here.
            fatalError("No UIWindowScene available to present ASWebAuthenticationSession")
        }

        return windowScene.windows.first(where: \.isKeyWindow)
            ?? windowScene.windows.first
            ?? UIWindow(windowScene: windowScene)
    }
}

// MARK: - CharacterSet

extension CharacterSet {
    /// A conservative character set safe for `application/x-www-form-urlencoded` values.
    fileprivate static let urlQueryValueAllowed: CharacterSet = {
        var allowed = CharacterSet.alphanumerics
        allowed.insert(charactersIn: "-._~")
        return allowed
    }()
}
