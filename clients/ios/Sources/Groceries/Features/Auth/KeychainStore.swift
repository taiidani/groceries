import Foundation
import Security

// MARK: - KeychainStore

/// A thin wrapper around the iOS Keychain Services API for storing and
/// retrieving the API Bearer token securely.
///
/// All operations are synchronous and performed on the calling thread.
/// Call sites should dispatch off the main thread if latency is a concern,
/// though Keychain operations are typically fast for small data items.
enum KeychainStore {

    private static let service = "JE539SF9V7.groceries"
    private static let tokenAccount = "api-token"

    // MARK: - Save

    /// Persists a string value in the Keychain, replacing any existing entry.
    ///
    /// - Parameters:
    ///   - value: The string to store (e.g. a Bearer token).
    ///   - account: The account key under which to store it.
    /// - Throws: `KeychainError.saveFailed` if the Keychain operation fails.
    static func save(_ value: String, account: String = tokenAccount) throws {
        guard let data = value.data(using: .utf8) else {
            throw KeychainError.encodingFailed
        }

        // Try an update first; if the item doesn't exist, add it.
        let query = baseQuery(account: account)
        let status = SecItemCopyMatching(query as CFDictionary, nil)

        if status == errSecSuccess || status == errSecInteractionNotAllowed {
            let attributes: [CFString: Any] = [
                kSecValueData: data,
                kSecAttrAccessible: kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly,
            ]
            let updateStatus = SecItemUpdate(query as CFDictionary, attributes as CFDictionary)
            if updateStatus != errSecSuccess {
                throw KeychainError.saveFailed(updateStatus)
            }
        } else if status == errSecItemNotFound {
            var addQuery = baseQuery(account: account)
            addQuery[kSecValueData] = data
            addQuery[kSecAttrAccessible] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
            let addStatus = SecItemAdd(addQuery as CFDictionary, nil)
            if addStatus != errSecSuccess {
                throw KeychainError.saveFailed(addStatus)
            }
        } else {
            throw KeychainError.saveFailed(status)
        }
    }

    // MARK: - Load

    /// Retrieves a previously stored string value from the Keychain.
    ///
    /// - Parameter account: The account key to look up.
    /// - Returns: The stored string, or `nil` if no entry was found.
    /// - Throws: `KeychainError.loadFailed` if an unexpected Keychain error occurs.
    static func load(account: String = tokenAccount) throws -> String? {
        var query = baseQuery(account: account)
        query[kSecReturnData] = true
        query[kSecMatchLimit] = kSecMatchLimitOne

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        switch status {
        case errSecSuccess:
            guard let data = result as? Data,
                let value = String(data: data, encoding: .utf8)
            else {
                throw KeychainError.decodingFailed
            }
            return value
        case errSecItemNotFound:
            return nil
        default:
            throw KeychainError.loadFailed(status)
        }
    }

    // MARK: - Delete

    /// Removes a stored value from the Keychain.
    ///
    /// Silently succeeds if no entry exists for `account`.
    ///
    /// - Parameter account: The account key to delete.
    /// - Throws: `KeychainError.deleteFailed` if an unexpected Keychain error occurs.
    static func delete(account: String = tokenAccount) throws {
        let query = baseQuery(account: account)
        let status = SecItemDelete(query as CFDictionary)
        if status != errSecSuccess && status != errSecItemNotFound {
            throw KeychainError.deleteFailed(status)
        }
    }

    // MARK: - Convenience token accessors

    /// Stores the API Bearer token.
    static func saveToken(_ token: String) throws {
        try save(token, account: tokenAccount)
    }

    /// Loads the API Bearer token, or `nil` if none is stored.
    static func loadToken() throws -> String? {
        try load(account: tokenAccount)
    }

    /// Removes the API Bearer token.
    static func deleteToken() throws {
        try delete(account: tokenAccount)
    }

    // MARK: - Private helpers

    private static func baseQuery(account: String) -> [CFString: Any] {
        [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: account,
        ]
    }
}

// MARK: - KeychainError

enum KeychainError: Error, LocalizedError {
    case encodingFailed
    case decodingFailed
    case saveFailed(OSStatus)
    case loadFailed(OSStatus)
    case deleteFailed(OSStatus)

    var errorDescription: String? {
        switch self {
        case .encodingFailed:
            return "Could not encode value for Keychain storage."
        case .decodingFailed:
            return "Could not decode value retrieved from Keychain."
        case .saveFailed(let status):
            return "Keychain save failed with status \(status)."
        case .loadFailed(let status):
            return "Keychain load failed with status \(status)."
        case .deleteFailed(let status):
            return "Keychain delete failed with status \(status)."
        }
    }
}
