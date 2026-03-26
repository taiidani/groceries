import SwiftUI
import GroceriesAPI

enum ItemsViewRoute: Hashable {
    case editor(itemID: Int)
}

enum ItemMembershipToggleContext {
    case listRows
    case addItemForm
    case editor
}

enum ItemMembershipToggleAccess {
    static func isAvailable(in context: ItemMembershipToggleContext) -> Bool {
        context == .editor
    }
}

enum ItemsViewAccessibility {
    static let searchFieldLabel = "Item search"
    static let inListOnlyToggleLabel = "In List only"
    static let addItemButtonLabel = "Add item"
}

enum ItemsViewUX {
    static func shouldUsePathDrivenNavigation() -> Bool {
        false
    }

    static func editorRoute(for item: Item) -> ItemsViewRoute {
        .editor(itemID: item.id)
    }

    static func editorItem(for route: ItemsViewRoute, items: [Item]) -> Item? {
        switch route {
        case .editor(let itemID):
            return items.first(where: { $0.id == itemID })
        }
    }

    static func addSheetInteractiveDismissDisabled(isAdding: Bool) -> Bool {
        isAdding
    }

    static func shouldShowRetryAffordance(
        isLoading: Bool,
        filteredItems: [Item],
        loadErrorMessage: String?,
        mutationErrorMessage _: String?
    ) -> Bool {
        !isLoading && filteredItems.isEmpty && loadErrorMessage != nil
    }

    static func performRetry(using action: () async -> Void) async {
        await action()
    }
}

struct ItemsView: View {
    let apiClient: GroceriesAPIClient

    @State private var viewModel: ItemsViewModel
    @State private var isPresentingAddItem = false
    init(apiClient: GroceriesAPIClient) {
        self.apiClient = apiClient
        _viewModel = State(initialValue: ItemsViewModel(api: apiClient))
    }

    var body: some View {
        NavigationStack {
            ZStack {
                AppBackground()

                List {
                    Section {
                        TextField(
                            "Search items",
                            text: Binding(
                                get: { viewModel.searchText },
                                set: { viewModel.searchText = $0 }
                            )
                        )
                        .accessibilityLabel(ItemsViewAccessibility.searchFieldLabel)

                        Toggle(
                            "In List only",
                            isOn: Binding(
                                get: { viewModel.inListOnly },
                                set: { viewModel.inListOnly = $0 }
                            )
                        )
                        .accessibilityLabel(ItemsViewAccessibility.inListOnlyToggleLabel)
                    }

                    if viewModel.filteredItems.isEmpty && !viewModel.isLoading {
                        Section {
                            if ItemsViewUX.shouldShowRetryAffordance(
                                isLoading: viewModel.isLoading,
                                filteredItems: viewModel.filteredItems,
                                loadErrorMessage: viewModel.loadErrorMessage,
                                mutationErrorMessage: viewModel.mutationErrorMessage
                            ) {
                                ContentUnavailableView {
                                    Label("Unable to load items", systemImage: "exclamationmark.triangle")
                                } description: {
                                    Text("Please try again.")
                                } actions: {
                                    Button("Retry") {
                                        Task {
                                            await ItemsViewUX.performRetry(using: viewModel.retryLoad)
                                        }
                                    }
                                }
                            } else {
                                ContentUnavailableView(
                                    "No items",
                                    systemImage: "square.grid.2x2",
                                    description: Text("Try adjusting search or add a new item.")
                                )
                            }
                        }
                    } else {
                        Section {
                            ForEach(viewModel.filteredItems) { item in
                                NavigationLink {
                                    ItemEditorView(item: item, viewModel: viewModel)
                                } label: {
                                    VStack(alignment: .leading, spacing: 4) {
                                        Text(item.name)
                                            .foregroundStyle(.white)
                                        Text(item.categoryName)
                                            .font(.caption)
                                            .foregroundStyle(.white.opacity(0.75))
                                    }
                                    .accessibilityElement(children: .combine)
                                    .accessibilityLabel(item.list == nil ? "\(item.name), \(item.categoryName)" : "\(item.name), \(item.categoryName), in list")
                                }
                            }
                        }
                    }
                }
                .listStyle(.insetGrouped)
                .scrollContentBackground(.hidden)
                .environment(\.colorScheme, .dark)
                .refreshable {
                    await viewModel.refresh()
                }
            }
            .navigationTitle("Items")
            .navigationBarTitleDisplayMode(.inline)
            .toolbarColorScheme(.dark, for: .navigationBar)
            .toolbarBackground(.clear, for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        isPresentingAddItem = true
                    } label: {
                        Label("Add Item", systemImage: "plus")
                    }
                    .accessibilityLabel(ItemsViewAccessibility.addItemButtonLabel)
                }
            }
            .task { await viewModel.load() }
            .sheet(isPresented: $isPresentingAddItem) {
                AddItemView(viewModel: viewModel)
                    .interactiveDismissDisabled(ItemsViewUX.addSheetInteractiveDismissDisabled(isAdding: viewModel.isAdding))
            }
        }
    }
}
