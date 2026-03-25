import SwiftUI
import GroceriesAPI

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
                    .disabled(viewModel.isAddCategoryPickerDisabled)
                    .accessibilityLabel("Item category")

                    TextField("Item name", text: $name)
                        .disabled(viewModel.isAddNameFieldDisabled)
                        .accessibilityLabel("Item name")
                }
            }
            .navigationTitle("Add Item")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                    .disabled(viewModel.isAdding)
                    .accessibilityLabel("Cancel add item")
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
                    .disabled(viewModel.isAddButtonDisabled(name: name, categoryID: selectedCategoryID))
                    .accessibilityLabel("Save item")
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
