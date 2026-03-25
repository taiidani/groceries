import SwiftUI
import GroceriesAPI

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
                        .accessibilityLabel("Item search")

                        Toggle(
                            "In List only",
                            isOn: Binding(
                                get: { viewModel.inListOnly },
                                set: { viewModel.inListOnly = $0 }
                            )
                        )
                        .accessibilityLabel("In List only")
                    }

                    if viewModel.filteredItems.isEmpty && !viewModel.isLoading {
                        Section {
                            ContentUnavailableView(
                                "No items",
                                systemImage: "square.grid.2x2",
                                description: Text("Try adjusting search or add a new item.")
                            )
                        }
                    } else {
                        Section {
                            ForEach(viewModel.filteredItems) { item in
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
                .listStyle(.insetGrouped)
                .scrollContentBackground(.hidden)
                .environment(\.colorScheme, .dark)
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
                    .accessibilityLabel("Add item")
                }
            }
            .task { await viewModel.load() }
            .sheet(isPresented: $isPresentingAddItem) {
                AddItemView(viewModel: viewModel)
            }
        }
    }
}
