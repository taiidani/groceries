# Groceries — iOS Client

A native SwiftUI shopping list app for iOS 26+, built with [Tuist](https://tuist.dev).

## Requirements

| Tool | Version |
|---|---|
| Xcode | 26.0+ |
| iOS Deployment Target | 26.0 |
| Swift | 6.0 |
| Tuist | 4.x (managed via `mise`) |

> **Note:** You need Xcode installed even if you edit source files in Zed, because
> Tuist delegates to Xcode's SDK toolchain for building and running on simulator/device.

## Project Structure

```
clients/ios/
├── Project.swift              # Tuist project manifest (edit this to add targets/files)
├── Tuist/
│   ├── Config.swift           # Tuist global configuration
│   └── Package.swift          # External Swift Package Manager dependencies
├── Sources/
│   ├── Groceries/             # App target
│   │   ├── GroceriesApp.swift         # @main entry point + RootView
│   │   ├── Resources/                 # Asset catalog, localisation strings
│   │   └── Features/
│   │       ├── Auth/
│   │       │   ├── AuthViewModel.swift
│   │       │   ├── LoginView.swift
│   │       │   └── KeychainStore.swift
│   │       ├── Navigation/
│   │       │   └── AppTabsView.swift
│   │       ├── Items/
│   │       │   ├── ItemsViewModel.swift
│   │       │   ├── ItemsView.swift
│   │       │   ├── AddItemView.swift
│   │       │   └── ItemEditorView.swift
│   │       └── ShoppingList/
│   │           ├── ShoppingListViewModel.swift
│   │           └── ShoppingListView.swift
│   └── GroceriesAPI/          # Framework target — API client & models
│       ├── GroceriesAPIClient.swift
│       ├── Models.swift
│       ├── AuthEndpoints.swift
│       ├── ListEndpoints.swift
│       └── ItemEndpoints.swift
└── Tests/
    ├── GroceriesAPITests/     # Unit tests for the API layer
    └── GroceriesTests/        # Unit tests for app features/view models
```

> **Xcode project files (`.xcodeproj`, `.xcworkspace`, `Derived/`) are git-ignored.**
> They are generated on demand by Tuist and should never be committed.

## Day-to-day Workflow

### 1. Generate the Xcode project

Run this once after cloning, and again whenever you add or remove source files,
targets, or dependencies in `Project.swift`:

```bash
cd clients/ios
tuist generate
```

This produces `Groceries.xcodeproj` locally (git-ignored).

### 2. Edit source files

Open `.swift` files directly in Zed (or your editor of choice). You do **not**
need to open Xcode just to edit code.

```bash
# From the repo root:
zed clients/ios/Sources/
```

### 3. Build & run

Open the generated project in Xcode to run on a simulator or device:

```bash
open Groceries.xcodeproj
```

Then select a simulator and press **⌘R**.

Alternatively, build from the terminal (no GUI required):

```bash
tuist build
```

### 4. When you add a new file

1. Create the `.swift` file in the appropriate `Sources/` subdirectory.
2. Run `tuist generate` to update the Xcode project.
3. Continue editing in Zed; Xcode will pick up the change automatically if it
   is already open.

### 5. Adding a Swift Package dependency

1. Add the dependency to `Tuist/Package.swift`.
2. Reference it in `Project.swift` under the relevant target's `dependencies:` array.
3. Run `tuist install` then `tuist generate`.

## Configuration

### API Base URL

The app reads `API_BASE_URL` from `Info.plist` at launch, falling back to
`http://localhost:3000`. To point a build scheme at a different server:

1. In Xcode, open **Product → Scheme → Edit Scheme…**
2. Under **Run → Arguments → Environment Variables**, add:
   ```
   API_BASE_URL = https://your-server.example.com
   ```

Or bake it in permanently by adding an `infoPlist` key in `Project.swift`:

```swift
"API_BASE_URL": "https://your-server.example.com"
```

### Apple Developer Team

The Team ID is `JE539SF9V7` and is set in `Project.swift`. Signing is handled
automatically by Xcode when you select a real device as the run destination.

## Architecture

The app follows a straightforward MVVM pattern using Swift's `@Observable` macro
(iOS 17+) and structured concurrency (`async`/`await`).

```
GroceriesApp (@main)
└── RootView               — routes between Login and main content
    ├── LoginView          — username/password form
    │   └── AuthViewModel  — owns GroceriesAPIClient, manages token lifecycle
    └── AppTabsView        — authenticated shell tab order: List -> Items -> Account
        ├── ShoppingListView      — list screen with deduped auto-refresh coordinator
        │   └── ShoppingListViewModel — list state and list mutations
        └── ItemsView             — item management screen
            └── ItemsViewModel    — item filtering, add/edit/delete, membership updates
```

### Key design decisions

| Decision | Rationale |
|---|---|
| `GroceriesAPI` is a separate framework target | Keeps networking/model code testable in isolation, independent of SwiftUI |
| `GroceriesAPIClient` is an `actor` | Ensures token mutation is data-race safe across async contexts |
| Bearer token stored in Keychain | Survives app restarts; more secure than `UserDefaults` |
| Dedicated `Features/Items` module | Keeps item CRUD and editor-only membership logic isolated from shopping list state |
| Notification-based cross-tab sync | Item membership changes publish an app event consumed by shopping list refresh coordination |
| No third-party dependencies (yet) | Reduces build complexity; revisit when pagination or richer offline support is needed |

## Running Tests

```bash
tuist test
```

Or in Xcode: **⌘U**.

## Future Work

- Offline caching/sync for list and item data
- Bulk item operations (multi-select delete/category move)
- Recipe support
- macOS Catalyst / dedicated macOS target
- Real-time updates via Server-Sent Events
