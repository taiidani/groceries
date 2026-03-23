import ProjectDescription

let project = Project(
    name: "Groceries",
    organizationName: "JE539SF9V7",
    options: .options(
        defaultKnownRegions: ["en"],
        developmentRegion: "en"
    ),
    targets: [
        .target(
            name: "GroceriesT",
            destinations: .iOS,
            product: .app,
            bundleId: "com.ryannixon.groceries",
            deploymentTargets: .iOS("26.0"),
            infoPlist: .extendingDefault(with: [
                "CFBundleDisplayName": "Groceries",
                "CFBundleShortVersionString": "1.0",
                "CFBundleVersion": "1",
                "CFBundleIconName": "AppIcon",
                "UIPrerenderedIcon": true,
                "UILaunchScreen": [:],
                "NSAppTransportSecurity": [
                    "NSAllowsLocalNetworking": true
                ],
                "API_BASE_URL": "$(API_BASE_URL)",
            ]),
            sources: ["Sources/Groceries/**"],
            resources: [
                .glob(
                    pattern: "Sources/Groceries/Resources/**",
                    excluding: ["Sources/Groceries/Resources/**/*.swift"])
            ],
            entitlements: .dictionary([
                "keychain-access-groups": .array([
                    .string("$(AppIdentifierPrefix)JE539SF9V7.groceries")
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
                            "API_BASE_URL": "http://localhost:3000",
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
                .target(name: "GroceriesT"),
                .target(name: "GroceriesAPI")
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
