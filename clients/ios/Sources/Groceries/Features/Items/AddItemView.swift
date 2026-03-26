import SwiftUI
import GroceriesAPI

enum AddItemViewAccessibility {
    static let categoryLabel = "Item category"
    static let nameLabel = "Item name"
    static let cancelButtonLabel = "Cancel add item"
    static let saveButtonLabel = "Save item"
    static let errorLabel = "Add item error"
}

enum AddItemViewUX {
    static func cancelDisabled(isAdding: Bool) -> Bool {
        isAdding
    }

    static func saveDisabled(isAdding: Bool, baseSaveDisabled: Bool) -> Bool {
        isAdding || baseSaveDisabled
    }

    static func categoryDisabled(isAdding: Bool) -> Bool {
        isAdding
    }

    static func nameDisabled(isAdding: Bool) -> Bool {
        isAdding
    }
}

struct AddItemView: View {
    @Environment(\.dismiss) private var dismiss

    let viewModel: ItemsViewModel

    @State private var selectedCategoryID: Int?
    @State private var name = ""

    var body: some View {
        NavigationStack {
            Form {
                Section("Details") {
                    Picker("Category", selection: $selectedCategoryID) {
                        Text("Select category").tag(Optional<Int>.none)
                        ForEach(viewModel.categories) { category in
                            Text(category.name).tag(Optional(category.id))
                        }
                    }
                    .disabled(AddItemViewUX.categoryDisabled(isAdding: viewModel.isAdding))
                    .accessibilityLabel(AddItemViewAccessibility.categoryLabel)

                    TextField("Item name", text: $name)
                        .disabled(AddItemViewUX.nameDisabled(isAdding: viewModel.isAdding))
                        .accessibilityLabel(AddItemViewAccessibility.nameLabel)
                }

                if let errorMessage = viewModel.mutationErrorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .accessibilityLabel(AddItemViewAccessibility.errorLabel)
                            .accessibilityValue(errorMessage)
                    }
                }
            }
            .navigationTitle("Add Item")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                    .disabled(AddItemViewUX.cancelDisabled(isAdding: viewModel.isAdding))
                    .accessibilityLabel(AddItemViewAccessibility.cancelButtonLabel)
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        Task {
                            let success = await viewModel.addItem(name: name, categoryID: selectedCategoryID)
                            if success {
                                dismiss()
                            }
                        }
                    }
                    .disabled(
                        AddItemViewUX.saveDisabled(
                            isAdding: viewModel.isAdding,
                            baseSaveDisabled: viewModel.isAddButtonDisabled(
                                name: name,
                                categoryID: selectedCategoryID
                            )
                        )
                    )
                    .accessibilityLabel(AddItemViewAccessibility.saveButtonLabel)
                }
            }
        }
    }
}

#Preview {
    let apiClient = GroceriesAPIClient(baseURL: URL(string: "http://localhost:3000")!)
    let viewModel = ItemsViewModel(api: apiClient)
    AddItemView(viewModel: viewModel)
}
