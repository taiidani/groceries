import Foundation
import ProjectDescription

// Allows `mise run xcode` to point Debug builds at the Mac's LAN IP (so a
// physical device can reach the locally-running server) without editing this
// file. Falls back to localhost for the simulator / default `tuist generate`.
// Must be TUIST_-prefixed: Tuist only forwards env vars with that prefix into
// the manifest (Project.swift) process.
let debugAPIBaseURL = ProcessInfo.processInfo.environment["TUIST_API_BASE_URL"] ?? "http://localhost:3000"

// When set (by `mise run xcode`), the Debug build is installed as a
// separate app — its own bundle ID, display name, URL scheme, and Keychain
// access group — so it never overwrites or shares data with a "real"
// Groceries install that talks to production data.
//
// The bundle ID uses "local" rather than "dev": iOS permanently records a
// silent "local network prohibited" decision for a bundle ID the first time
// it attempts local-network access without NSLocalNetworkUsageDescription
// present, and that decision persists (with no Settings entry to undo it)
// across uninstall/reinstall. "com.ryannixon.groceries.dev" got poisoned
// this way during development before the key was added; if this identifier
// ever suffers the same fate, rename it again rather than debugging further.
let isDevBuild = ProcessInfo.processInfo.environment["TUIST_DEV_BUILD"] != nil
let appBundleId = isDevBuild ? "com.ryannixon.groceries.local" : "com.ryannixon.groceries"
let appDisplayName = isDevBuild ? "Groceries Dev" : "Groceries"
let keychainAccessGroup = isDevBuild ? "JE539SF9V7.groceries.local" : "JE539SF9V7.groceries"

let project = Project(
    name: "Groceries",
    organizationName: "JE539SF9V7",
    options: .options(
        defaultKnownRegions: ["en"],
        developmentRegion: "en"
    ),
    targets: [
        .target(
            name: "Aisle4",
            destinations: .iOS,
            product: .app,
            bundleId: appBundleId,
            deploymentTargets: .iOS("26.0"),
            infoPlist: .extendingDefault(
                with: [
                    "CFBundleDisplayName": .string(appDisplayName),
                    "CFBundleShortVersionString": "1.0",
                    "CFBundleVersion": "1",
                    "CFBundleIconName": "AppIcon",
                    "UIPrerenderedIcon": true,
                    "UILaunchScreen": [:],
                    "NSAppTransportSecurity": [
                        "NSAllowsLocalNetworking": true
                    ],
                    "API_BASE_URL": "$(API_BASE_URL)",
                    "CFBundleURLTypes": [
                        [
                            "CFBundleURLName": .string("\(appBundleId).auth"),
                            "CFBundleURLSchemes": [.string(appBundleId)],
                        ]
                    ],
                ]
                // Local Network access (and the permission prompt for it) is
                // only ever needed talking to a Mac's LAN IP during dev-build
                // testing; a real Release build always talks to production
                // over HTTPS, so these keys — and the runtime Bonjour probe
                // in `LocalNetworkPermission.swift`, which is itself gated by
                // `#if DEBUG` — are omitted entirely outside dev builds.
                .merging(
                    isDevBuild
                        ? [
                            "NSLocalNetworkUsageDescription":
                                "Groceries needs to connect to the API server on your local network.",
                            // A declared Bonjour service type (even one we don't
                            // actually advertise/discover anything meaningful
                            // with) is required to reliably trigger iOS's Local
                            // Network permission *prompt* — plain URLSession
                            // requests to a raw IP address alone often never
                            // surface it.
                            "NSBonjourServices": ["_http._tcp."],
                        ]
                        : [:]
                ) { _, new in new }
            ),
            sources: ["Sources/Groceries/**"],
            resources: [
                .glob(
                    pattern: "Sources/Groceries/Resources/**",
                    excluding: ["Sources/Groceries/Resources/**/*.swift"])
            ],
            entitlements: .dictionary([
                "keychain-access-groups": .array([
                    .string("$(AppIdentifierPrefix)\(keychainAccessGroup)")
                ])
            ]),
            dependencies: [
                .target(name: "GroceriesAPI")
            ],
            settings: .settings(
                base: [
                    "DEVELOPMENT_TEAM": "JE539SF9V7",
                    "SWIFT_VERSION": "6.0",
                    "IPHONEOS_DEPLOYMENT_TARGET": "26.0",
                    "CODE_SIGN_ALLOW_ENTITLEMENTS_MODIFICATION": "YES",
                    // Use the single 1024x1024 universal icon — prevents actool from
                    // resizing it down to legacy sizes and compositing over white.
                    "ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon",
                    "ASSETCATALOG_COMPILER_INCLUDE_ALL_APPICON_ASSETS": "YES",
                    // Disable the legacy icon sizes that strip transparency.
                    "ASSETCATALOG_COMPILER_SKIP_APP_STORE_DEPLOYMENT": "YES",
                ],
                configurations: [
                    .debug(
                        name: "Debug",
                        settings: [
                            "SWIFT_ACTIVE_COMPILATION_CONDITIONS": "DEBUG",
                            "API_BASE_URL": SettingValue(stringLiteral: debugAPIBaseURL),
                        ]),
                    .release(
                        name: "Release",
                        settings: [
                            "API_BASE_URL": "https://groceries.taiidani.com"
                        ]),
                ]
            )
        ),
        .target(
            name: "GroceriesAPI",
            destinations: .iOS,
            product: .framework,
            bundleId: "com.ryannixon.groceries.api",
            deploymentTargets: .iOS("26.0"),
            infoPlist: .default,
            sources: ["Sources/GroceriesAPI/**"],
            dependencies: [],
            settings: .settings(
                base: [
                    "DEVELOPMENT_TEAM": "JE539SF9V7",
                    "SWIFT_VERSION": "6.0",
                    "IPHONEOS_DEPLOYMENT_TARGET": "26.0",
                ]
            )
        ),
        .target(
            name: "GroceriesAPITests",
            destinations: .iOS,
            product: .unitTests,
            bundleId: "com.ryannixon.groceries.api.tests",
            deploymentTargets: .iOS("26.0"),
            infoPlist: .default,
            sources: ["Tests/GroceriesAPITests/**"],
            dependencies: [
                .target(name: "GroceriesAPI")
            ],
            settings: .settings(
                base: [
                    "DEVELOPMENT_TEAM": "JE539SF9V7",
                    "SWIFT_VERSION": "6.0",
                ]
            )
        ),
        .target(
            name: "GroceriesTests",
            destinations: .iOS,
            product: .unitTests,
            bundleId: "com.ryannixon.groceries.tests",
            deploymentTargets: .iOS("26.0"),
            infoPlist: .default,
            sources: ["Tests/GroceriesTests/**"],
            dependencies: [
                .target(name: "Aisle4"),
                .target(name: "GroceriesAPI"),
            ],
            settings: .settings(
                base: [
                    "DEVELOPMENT_TEAM": "JE539SF9V7",
                    "SWIFT_VERSION": "6.0",
                ]
            )
        ),
    ]
)
