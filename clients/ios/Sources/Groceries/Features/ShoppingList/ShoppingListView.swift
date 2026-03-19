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
/// - Toolbar "Finish Shopping" button only visible when there are done items.
/// - Empty state with a friendly illustration and message.
/// - Inline error banner rather than disruptive alerts.
struct ShoppingListView: View {

    // MARK: - Dependencies

    @Environment(AuthViewModel.self) private var authViewModel

    // MARK: - View model

    @State private var viewModel: ShoppingListViewModel

    // MARK: - Local UI state

    @State private var showFinishConfirmation: Bool = false

    // MARK: - Init

    init(apiClient: GroceriesAPIClient) {
        _viewModel = State(initialValue: ShoppingListViewModel(apiClient: apiClient))
    }

    // MARK: - Computed

    // MARK: - Body

    var body: some View {
        NavigationStack {
            ZStack {
                AppBackground()

                if viewModel.isLoading && viewModel.isEmpty {
                    loadingView
                } else if viewModel.storeGroups.isEmpty && !viewModel.isLoading {
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
            .safeAreaInset(edge: .bottom) {
                // Error banner pinned above the home indicator.
                if let message = viewModel.errorMessage {
                    errorBanner(message: message)
                        .padding(.horizontal)
                        .padding(.bottom, 8)
                        .transition(.move(edge: .bottom).combined(with: .opacity))
                }
            }
            .animation(.easeInOut(duration: 0.25), value: viewModel.errorMessage)
            .confirmationDialog(
                "Finish Shopping?",
                isPresented: $showFinishConfirmation,
                titleVisibility: .visible
            ) {
                Button("Remove \(viewModel.totalDone) done item(s)", role: .destructive) {
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
        List {
            progressSection

            ForEach(viewModel.storeGroups) { store in
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
        }
        .listStyle(.insetGrouped)
        .scrollContentBackground(.hidden)
        .environment(\.colorScheme, .dark)
        .animation(.default, value: viewModel.storeGroups.map(\.id))
    }

    // MARK: - Progress section

    /// A compact progress summary at the top of the list.
    @ViewBuilder
    private var progressSection: some View {
        if viewModel.total > 0 {
            Section {
                ProgressSummaryView(
                    total: viewModel.total,
                    done: viewModel.totalDone
                )
            }
            .listRowInsets(.init(top: 8, leading: 16, bottom: 8, trailing: 16))
        }
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
        ToolbarItem(placement: .topBarTrailing) {
            if viewModel.hasDoneItems {
                Button {
                    showFinishConfirmation = true
                } label: {
                    Label("Finish Shopping", systemImage: "checkmark.circle")
                }
                .disabled(viewModel.isMutating)
                .accessibilityLabel(
                    "Finish Shopping — remove \(viewModel.totalDone) done item(s)"
                )
            }
        }

        ToolbarItem(placement: .topBarLeading) {
            if viewModel.isMutating {
                ProgressView()
                    .accessibilityLabel("Updating list…")
            }
        }
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
