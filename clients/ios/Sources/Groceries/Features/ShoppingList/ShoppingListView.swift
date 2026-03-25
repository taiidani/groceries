import GroceriesAPI
import SwiftUI

// MARK: - Background

/// The fruit pattern background with a dark scrim, shared across screens.
struct AppBackground: View {
    var body: some View {
        Image("BackgroundImage")
            .resizable(resizingMode: .tile)
            .overlay(Color.black.opacity(0.55))
            .ignoresSafeArea()
    }
}

// MARK: - ShoppingListView

/// The main shopping list screen.
///
/// Conforms to Apple's Human Interface Guidelines:
/// - Uses `List` with swipe actions for destructive operations.
/// - Groups items by category using `Section`.
/// - Pull-to-refresh via `.refreshable`.
/// - Global progress summary shown once for the whole shopping list.
/// - Empty state with a friendly illustration and message.
/// - Inline error banner rather than disruptive alerts.
struct ShoppingListView: View {

    // MARK: - Dependencies

    @Environment(AuthViewModel.self) private var authViewModel
    @Environment(\.dynamicTypeSize) private var dynamicTypeSize

    // MARK: - View model

    @State private var viewModel: ShoppingListViewModel

    // MARK: - Local UI state

    @State private var showFinishConfirmation: Bool = false
    @State private var selectedStoreID: Int?

    // MARK: - Init

    init(apiClient: GroceriesAPIClient) {
        _viewModel = State(initialValue: ShoppingListViewModel(apiClient: apiClient))
    }

    // MARK: - Computed

    private var availableStoreIDs: [Int] {
        viewModel.nonEmptyStoreGroups.map(\.id)
    }

    // MARK: - Body

    var body: some View {
        NavigationStack {
            ZStack {
                AppBackground()

                if viewModel.isLoading && viewModel.isEmpty {
                    loadingView
                } else if viewModel.nonEmptyStoreGroups.isEmpty && !viewModel.isLoading {
                    emptyStateView
                } else {
                    listView
                }
            }
            .navigationTitle("Shopping List")
            .navigationBarTitleDisplayMode(.large)
            .toolbarColorScheme(.dark, for: .navigationBar)
            .toolbarBackground(.clear, for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
            .toolbar { toolbarContent }
            .refreshable { await viewModel.refresh() }
            .task { await viewModel.load() }
            .onAppear { reconcileStoreSelection() }
            .onChange(of: availableStoreIDs) { _, _ in
                reconcileStoreSelection()
            }
            .safeAreaInset(edge: .bottom) {
                VStack(spacing: 8) {
                    AddItemBar(
                        search: { viewModel.searchItems(query: $0) },
                        onAdd: { itemID, name, quantity in
                            try await viewModel.addItem(itemID: itemID, name: name, quantity: quantity)
                        }
                    )

                    // Error banner pinned above the home indicator.
                    if let message = viewModel.errorMessage {
                        errorBanner(message: message)
                            .transition(.move(edge: .bottom).combined(with: .opacity))
                    }
                }
                .padding(.horizontal)
                .padding(.bottom, 8)
            }
            .animation(.easeInOut(duration: 0.25), value: viewModel.errorMessage)
            .confirmationDialog(
                "Clear Done Items?",
                isPresented: $showFinishConfirmation,
                titleVisibility: .visible
            ) {
                Button(
                    "Remove \(viewModel.totalDone) done \(doneItemWord(for: viewModel.totalDone))",
                    role: .destructive
                ) {
                    Task { await viewModel.finishShopping() }
                }
                Button("Cancel", role: .cancel) {}
            } message: {
                Text("All items marked as done will be removed from the list.")
            }
        }
    }

    // MARK: - List view

    private var listView: some View {
        VStack(spacing: 12) {
            listHeader

            storeChipsView

            TabView(selection: $selectedStoreID) {
                ForEach(viewModel.nonEmptyStoreGroups) { store in
                    storeList(store: store)
                        .tag(Optional(store.id))
                }
            }
            .tabViewStyle(.page(indexDisplayMode: .never))
            .indexViewStyle(.page(backgroundDisplayMode: .never))
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .animation(.default, value: availableStoreIDs)
    }

    @ViewBuilder
    private var listHeader: some View {
        if viewModel.total > 0 {
            Group {
                if dynamicTypeSize.isAccessibilitySize {
                    VStack(alignment: .leading, spacing: 10) {
                        headerProgressSummary
                        clearDoneButton
                    }
                } else {
                    HStack(alignment: .center, spacing: 12) {
                        headerProgressSummary
                        clearDoneButton
                    }
                }
            }
            .padding(.horizontal, 16)
        }
    }

    private var headerProgressSummary: some View {
        ProgressSummaryView(
            total: viewModel.total,
            done: viewModel.totalDone
        )
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    private var clearDoneButton: some View {
        if viewModel.hasDoneItems {
            Button {
                showFinishConfirmation = true
            } label: {
                Label("Clear Done", systemImage: "checkmark.circle")
            }
            .buttonStyle(.borderedProminent)
            .disabled(viewModel.isMutating)
            .accessibilityLabel(
                "Clear done items — remove \(viewModel.totalDone) done \(doneItemWord(for: viewModel.totalDone))"
            )
        }
    }

    private var storeChipsView: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 10) {
                ForEach(viewModel.nonEmptyStoreGroups) { store in
                    let isSelected = selectedStoreID == store.id
                    let isComplete = viewModel.isStoreComplete(storeID: store.id)

                    Button {
                        selectedStoreID = store.id
                    } label: {
                        HStack(spacing: 6) {
                            Text(store.name)
                                .lineLimit(1)

                            if isComplete {
                                Image(systemName: "checkmark.circle.fill")
                                    .font(.caption.weight(.bold))
                            }
                        }
                        .font(.subheadline.weight(.semibold))
                        .foregroundStyle(isSelected ? Color.black : Color.white)
                        .padding(.vertical, 8)
                        .padding(.horizontal, 12)
                        .background(
                            Capsule(style: .continuous)
                                .fill(isSelected ? Color.white : Color.white.opacity(0.15))
                        )
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel(storeChipAccessibilityLabel(for: store, isComplete: isComplete))
                    .accessibilityAddTraits(isSelected ? [.isButton, .isSelected] : .isButton)
                }
            }
            .padding(.horizontal, 16)
        }
    }

    private func storeList(store: StoreGroup) -> some View {
        List {
            Section {
                ForEach(store.categories) { category in
                    // Category name as a row header within the store section.
                    Text(category.name)
                        .font(.subheadline.weight(.semibold))
                        .foregroundStyle(.white.opacity(0.6))
                        .listRowBackground(Color.clear)
                        .listRowInsets(.init(top: 8, leading: 16, bottom: 2, trailing: 16))

                    ForEach(category.items) { item in
                        ShoppingListRow(
                            item: item,
                            isMutating: viewModel.mutatingItemIDs.contains(item.itemID),
                            onToggleDone: {
                                Task { await viewModel.toggleDone(for: item) }
                            }
                        )
                        .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                            Button(role: .destructive) {
                                Task { await viewModel.remove(item: item) }
                            } label: {
                                Label("Remove", systemImage: "trash")
                            }
                        }
                    }
                }
            } header: {
                Text(store.name)
                    .foregroundStyle(.white)
                    .textCase(nil)
                    .font(.headline)
            }
        }
        .listStyle(.insetGrouped)
        .scrollContentBackground(.hidden)
        .environment(\.colorScheme, .dark)
    }

    // MARK: - Loading view

    private var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .scaleEffect(1.4)
            Text("Loading your list…")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
    }

    // MARK: - Empty state

    private var emptyStateView: some View {
        ContentUnavailableView(
            "Your list is empty",
            systemImage: "cart",
            description: Text("Add items to your shopping list and they'll appear here.")
        )
    }

    // MARK: - Error banner

    private func errorBanner(message: String) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: 8) {
            Image(systemName: "exclamationmark.circle.fill")
                .foregroundStyle(.red)
                .accessibilityHidden(true)

            Text(message)
                .font(.subheadline)
                .foregroundStyle(.red)
                .fixedSize(horizontal: false, vertical: true)

            Spacer()

            Button {
                viewModel.clearError()
            } label: {
                Image(systemName: "xmark")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.secondary)
            }
            .accessibilityLabel("Dismiss error")
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10, style: .continuous)
                .fill(Color.red.opacity(0.1))
                .shadow(color: .black.opacity(0.06), radius: 4, y: 2)
        )
    }

    // MARK: - Toolbar

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItem(placement: .topBarLeading) {
            if viewModel.isMutating {
                ProgressView()
                    .accessibilityLabel("Updating list…")
            }
        }
    }

    private func reconcileStoreSelection() {
        selectedStoreID = StoreSelectionReconciler.reconcile(
            current: selectedStoreID,
            availableStoreIDs: availableStoreIDs
        )
    }

    private func storeChipAccessibilityLabel(for store: StoreGroup, isComplete: Bool) -> String {
        let totals = viewModel.storeTotals(storeID: store.id)
        let itemWord = totals.total == 1 ? "item" : "items"
        if isComplete {
            return "\(store.name), \(totals.total) \(itemWord), all done"
        }
        return "\(store.name), \(totals.total) \(itemWord)"
    }

    private func doneItemWord(for count: Int) -> String {
        count == 1 ? "item" : "items"
    }
}

// MARK: - ShoppingListRow

/// A single row in the shopping list.
///
/// Tapping the leading checkmark area toggles the done state.
/// The row dims when a mutation for this item is in flight.
private struct ShoppingListRow: View {

    let item: ListItem
    let isMutating: Bool
    let onToggleDone: () -> Void

    var body: some View {
        Button(action: onToggleDone) {
            HStack(spacing: 14) {
                // Done indicator
                Image(systemName: item.done ? "checkmark.circle.fill" : "circle")
                    .font(.title3)
                    .foregroundStyle(item.done ? Color.accentColor : Color.secondary)
                    .animation(.easeInOut(duration: 0.15), value: item.done)

                VStack(alignment: .leading, spacing: 2) {
                    Text(item.itemName)
                        .font(.body)
                        .strikethrough(item.done, color: .secondary)
                        .foregroundStyle(item.done ? .secondary : .primary)
                        .animation(.easeInOut(duration: 0.15), value: item.done)

                    if !item.quantity.isEmpty {
                        Text(item.quantity)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Spacer()

                if isMutating {
                    ProgressView()
                        .scaleEffect(0.8)
                }
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .disabled(isMutating)
        .opacity(isMutating ? 0.6 : 1)
        .animation(.easeInOut(duration: 0.15), value: isMutating)
        .accessibilityElement(children: .combine)
        .accessibilityLabel(accessibilityLabel)
        .accessibilityHint(item.done ? "Tap to mark as not done" : "Tap to mark as done")
        .accessibilityAddTraits(item.done ? [.isButton, .isSelected] : .isButton)
    }

    private var accessibilityLabel: String {
        var label = item.itemName
        if !item.quantity.isEmpty {
            label += ", \(item.quantity)"
        }
        if item.done {
            label += ", done"
        }
        return label
    }
}

// MARK: - ProgressSummaryView

/// A compact progress bar + label shown at the top of the list.
private struct ProgressSummaryView: View {
    let total: Int
    let done: Int

    private var progress: Double {
        total > 0 ? Double(done) / Double(total) : 0
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("\(done) of \(total) items done")
                    .font(.subheadline.weight(.medium))
                Spacer()
                Text("\(Int(progress * 100))%")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .monospacedDigit()
            }

            ProgressView(value: progress)
                .tint(done == total && total > 0 ? Color.green : Color.accentColor)
                .animation(.easeInOut(duration: 0.3), value: progress)
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(
            "\(done) of \(total) items done, \(Int(progress * 100)) percent complete")
    }
}

// MARK: - Previews

#Preview("Shopping List — items") {
    let vm = AuthViewModel(baseURL: URL(string: "http://localhost:3000")!)
    return ShoppingListView(apiClient: vm.apiClient)
        .environment(vm)
}

#Preview("Shopping List — empty") {
    let vm = AuthViewModel(baseURL: URL(string: "http://localhost:3000")!)
    return ShoppingListView(apiClient: vm.apiClient)
        .environment(vm)
}
