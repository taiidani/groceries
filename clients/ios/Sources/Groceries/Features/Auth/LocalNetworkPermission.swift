import Network

// MARK: - LocalNetworkPermission

/// Forces iOS to show its Local Network permission alert (if not already
/// decided) by performing a real, if trivial, Bonjour browse.
///
/// A plain `URLSession` request to a raw LAN IP address is often *not*
/// enough to trigger the system permission prompt at all — in practice it
/// can silently settle on a "prohibited" decision with no dialog ever
/// shown, and no way to grant access afterwards short of a device-wide
/// privacy reset. Only an actual Bonjour/`Network.framework` browse
/// reliably surfaces the alert. The declared service type here
/// (`_http._tcp.`, matching `NSBonjourServices` in `Project.swift`) doesn't
/// need to resolve to anything real — starting the browse is what matters.
enum LocalNetworkPermission {

    /// Starts a short-lived Bonjour browse and waits for it to reach a
    /// terminal state (or a safety timeout), then cancels it. Call once at
    /// app launch, before any other local-network traffic.
    @MainActor
    static func requestIfNeeded() async {
        let browser = NWBrowser(for: .bonjour(type: "_http._tcp.", domain: nil), using: .tcp)

        await withCheckedContinuation { (continuation: CheckedContinuation<Void, Never>) in
            let box = ResumeBox(browser: browser, continuation: continuation)

            browser.stateUpdateHandler = { state in
                switch state {
                case .ready, .failed, .cancelled:
                    Task { @MainActor in box.resume() }
                default:
                    break
                }
            }

            browser.start(queue: .main)

            // Safety net: the permission alert itself can take a moment to
            // be answered (or the browser may just sit in `.ready` without
            // further state changes), so don't block launch indefinitely.
            DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
                box.resume()
            }
        }
    }

    /// Ensures the continuation is resumed exactly once, regardless of
    /// which callback (state update vs. safety timeout) fires first.
    @MainActor
    private final class ResumeBox {
        private let browser: NWBrowser
        private var continuation: CheckedContinuation<Void, Never>?

        init(browser: NWBrowser, continuation: CheckedContinuation<Void, Never>) {
            self.browser = browser
            self.continuation = continuation
        }

        func resume() {
            guard let continuation else { return }
            self.continuation = nil
            browser.cancel()
            continuation.resume()
        }
    }
}
