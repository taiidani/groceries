import Foundation

final class MockURLProtocol: URLProtocol {
    typealias RequestHandler = @Sendable (URLRequest) throws -> (HTTPURLResponse, Data)

    private final class HandlerStorage: @unchecked Sendable {
        private let lock = NSLock()
        private var requestHandler: RequestHandler?

        func get() -> RequestHandler? {
            lock.lock()
            defer { lock.unlock() }
            return requestHandler
        }

        func set(_ handler: RequestHandler?) {
            lock.lock()
            requestHandler = handler
            lock.unlock()
        }
    }

    private static let handlerStorage = HandlerStorage()

    static func setRequestHandler(_ handler: @escaping RequestHandler) {
        handlerStorage.set(handler)
    }

    static func clearRequestHandler() {
        handlerStorage.set(nil)
    }

    static func requestBodyData(from request: URLRequest) -> Data? {
        if let body = request.httpBody {
            return body
        }

        guard let stream = request.httpBodyStream else {
            return nil
        }

        stream.open()
        defer { stream.close() }

        var data = Data()
        let bufferSize = 1024
        var buffer = [UInt8](repeating: 0, count: bufferSize)

        while stream.hasBytesAvailable {
            let bytesRead = stream.read(&buffer, maxLength: bufferSize)

            if bytesRead < 0 {
                return nil
            }

            if bytesRead == 0 {
                break
            }

            data.append(buffer, count: bytesRead)
        }

        return data
    }

    override class func canInit(with request: URLRequest) -> Bool {
        true
    }

    override class func canonicalRequest(for request: URLRequest) -> URLRequest {
        request
    }

    override func startLoading() {
        guard let handler = Self.handlerStorage.get() else {
            client?.urlProtocol(self, didFailWithError: NSError(domain: "MockURLProtocol", code: 0))
            return
        }

        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}
