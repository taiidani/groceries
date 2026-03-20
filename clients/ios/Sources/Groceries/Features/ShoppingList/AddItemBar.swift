import GroceriesAPI
import SwiftUI

struct AddItemBar: View {
    let search: (String) -> [Item]
    let onAdd: (_ itemID: Int?, _ name: String?, _ quantity: String) async throws -> Void

    @State private var mode: Mode = .idle
    @State private var query: String = ""
    @State private var quantity: String = ""
    @State private var isAdding: Bool = false
    @State private var addErrorMessage: String?

    @FocusState private var focusedField: FocusField?

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            switch mode {
            case .idle:
                idleView

            case .searching:
                searchingView

            case .quantity(let selection):
                quantityView(selection: selection)
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 18, style: .continuous)
                .fill(Color.black.opacity(0.38))
        )
        .overlay(
            RoundedRectangle(cornerRadius: 18, style: .continuous)
                .strokeBorder(Color.white.opacity(0.08), lineWidth: 1)
        )
    }

    private var idleView: some View {
        Button {
            mode = .searching
            focusedField = .search
        } label: {
            HStack(spacing: 8) {
                Image(systemName: "plus.circle.fill")
                    .foregroundStyle(.green)
                Text("Add an item...")
                    .foregroundStyle(.white.opacity(0.85))
                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
            .background(
                Capsule(style: .continuous)
                    .fill(Color.white.opacity(0.12))
            )
        }
        .buttonStyle(.plain)
    }

    private var searchingView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                TextField("Search items", text: $query)
                    .textInputAutocapitalization(.words)
                    .autocorrectionDisabled(false)
                    .focused($focusedField, equals: .search)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 10)
                    .background(
                        RoundedRectangle(cornerRadius: 12, style: .continuous)
                            .fill(Color.white.opacity(0.12))
                    )
                    .accessibilityLabel("Search items")

                Button("Cancel") {
                    resetToIdle()
                }
                .foregroundStyle(.white.opacity(0.85))
            }

            if !searchResults.isEmpty || !trimmedQuery.isEmpty {
                ScrollView {
                    VStack(spacing: 6) {
                        ForEach(searchResults) { item in
                            resultRow(
                                title: item.name,
                                subtitle: item.categoryName,
                                accessibilityLabel: "\(item.name), \(item.categoryName)"
                            ) {
                                selectExisting(item)
                            }
                        }

                        if !trimmedQuery.isEmpty {
                            resultRow(
                                title: "Add \"\(trimmedQuery)\" as new item",
                                subtitle: nil,
                                accessibilityLabel: "Add \(trimmedQuery) as new item"
                            ) {
                                selectNew(name: trimmedQuery)
                            }
                        }
                    }
                }
                .frame(maxHeight: 180)
            }
        }
    }

    private func quantityView(selection: Selection) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                Text(selection.label)
                    .lineLimit(1)
                    .font(.subheadline.weight(.semibold))
                    .foregroundStyle(.white)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 8)
                    .background(
                        Capsule(style: .continuous)
                            .fill(Color.white.opacity(0.16))
                    )

                Spacer(minLength: 0)

                Button("Cancel") {
                    resetToIdle()
                }
                .foregroundStyle(.white.opacity(0.85))
                .disabled(isAdding)
            }

            HStack(spacing: 8) {
                TextField("Qty", text: $quantity)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled(true)
                    .submitLabel(.done)
                    .onSubmit {
                        Task {
                            await performAdd(selection: selection)
                        }
                    }
                    .focused($focusedField, equals: .quantity)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 10)
                    .background(
                        RoundedRectangle(cornerRadius: 12, style: .continuous)
                            .fill(Color.white.opacity(0.12))
                    )
                    .disabled(isAdding)

                Button {
                    Task {
                        await performAdd(selection: selection)
                    }
                } label: {
                    if isAdding {
                        ProgressView()
                            .tint(.white)
                            .frame(maxWidth: .infinity)
                    } else {
                        Text("Add")
                            .fontWeight(.semibold)
                            .frame(maxWidth: .infinity)
                    }
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)
                .disabled(isAdding)
                .accessibilityLabel("Add item to shopping list")
            }

            if let addErrorMessage {
                Text(addErrorMessage)
                    .font(.caption)
                    .foregroundStyle(.red.opacity(0.9))
                    .accessibilityLabel("Add item failed: \(addErrorMessage)")
            }
        }
    }

    private func resultRow(
        title: String,
        subtitle: String?,
        accessibilityLabel: String,
        action: @escaping () -> Void
    ) -> some View {
        Button(action: action) {
            HStack(spacing: 8) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .foregroundStyle(.white)
                        .multilineTextAlignment(.leading)

                    if let subtitle {
                        Text(subtitle)
                            .font(.caption)
                            .foregroundStyle(.white.opacity(0.62))
                    }
                }

                Spacer(minLength: 0)
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 8)
            .background(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .fill(Color.white.opacity(0.08))
            )
        }
        .buttonStyle(.plain)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(accessibilityLabel)
    }

    private var searchResults: [Item] {
        search(trimmedQuery)
    }

    private var trimmedQuery: String {
        query.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func selectExisting(_ item: Item) {
        addErrorMessage = nil
        mode = .quantity(
            Selection(
                itemID: item.id,
                label: "\(item.name) \(item.categoryName)",
                trimmedName: nil
            ))
        quantity = ""
        focusedField = .quantity
    }

    private func selectNew(name: String) {
        addErrorMessage = nil
        mode = .quantity(
            Selection(
                itemID: nil,
                label: "\"\(name)\"",
                trimmedName: name
            ))
        quantity = ""
        focusedField = .quantity
    }

    private func performAdd(selection: Selection) async {
        guard !isAdding else { return }

        isAdding = true
        defer { isAdding = false }

        do {
            try await onAdd(selection.itemID, selection.trimmedName, quantity)
            addErrorMessage = nil
            resetToIdle()
        } catch {
            let message = error.localizedDescription.trimmingCharacters(in: .whitespacesAndNewlines)
            addErrorMessage = message.isEmpty ? "Could not add item. Please try again." : message
        }
    }

    private func resetToIdle() {
        mode = .idle
        query = ""
        quantity = ""
        addErrorMessage = nil
        focusedField = nil
    }
}

private enum Mode {
    case idle
    case searching
    case quantity(Selection)
}

private struct Selection {
    let itemID: Int?
    let label: String
    let trimmedName: String?
}

private enum FocusField: Hashable {
    case search
    case quantity
}

#Preview("Add Item Bar") {
    ZStack {
        AppBackground()
        AddItemBar(
            search: { _ in [] },
            onAdd: { _, _, _ in
                try await Task.sleep(for: .milliseconds(700))
            }
        )
        .padding()
    }
}
