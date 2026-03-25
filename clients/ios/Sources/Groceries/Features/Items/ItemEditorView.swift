import SwiftUI
import GroceriesAPI

enum ItemEditorViewAccessibility {
    static let categoryLabel = "Edit item category"
    static let nameLabel = "Edit item name"
    static let membershipToggleLabel = "Include item in shopping list"
    static let cancelButtonLabel = "Cancel edit item"
    static let saveButtonLabel = "Save item changes"
    static let deleteButtonLabel = "Delete item"
    static let deleteConfirmButtonLabel = "Confirm delete item"
    static let errorLabel = "Edit item error"
}

enum ItemEditorViewUX {
    static func cancelDisabled(isMutationInFlight: Bool) -> Bool {
        isMutationInFlight
    }

    static func saveDisabled(isMutationInFlight: Bool, baseSaveDisabled: Bool) -> Bool {
        isMutationInFlight || baseSaveDisabled
    }

    static func nameDisabled(isMutationInFlight: Bool) -> Bool {
        isMutationInFlight
    }

    static func categoryDisabled(isMutationInFlight: Bool) -> Bool {
        isMutationInFlight
    }

    static func membershipToggleDisabled(isMutationInFlight: Bool) -> Bool {
        isMutationInFlight
    }

    static func deleteDisabled(isMutationInFlight: Bool) -> Bool {
        isMutationInFlight
    }

    struct MembershipToggleMutation {
        let visibleValue: Bool
        let requestedValue: Bool
    }

    static func membershipToggleInitialState(isInList: Bool) -> MembershipToggleMutation {
        MembershipToggleMutation(visibleValue: isInList, requestedValue: isInList)
    }

    static func membershipToggleBeginMutation(currentVisibleValue: Bool, requestedValue: Bool) -> MembershipToggleMutation {
        MembershipToggleMutation(visibleValue: currentVisibleValue, requestedValue: requestedValue)
    }

    static func membershipToggleResolveMutation(
        currentVisibleValue: Bool,
        requestedValue: Bool,
        success: Bool
    ) -> Bool {
        if success {
            return requestedValue
        }
        return currentVisibleValue
    }

    static func membershipToggleSyncExternalModelChange(
        currentVisibleValue: Bool,
        modelIsInList: Bool,
        isMutationInFlight: Bool
    ) -> Bool {
        guard !isMutationInFlight else {
            return currentVisibleValue
        }
        return modelIsInList
    }
}

struct ItemEditorView: View {
    @Environment(\.dismiss) private var dismiss

    let item: Item
    let viewModel: ItemsViewModel

    @State private var selectedCategoryID: Int?
    @State private var name: String
    @State private var isInList: Bool
    @State private var isShowingDeleteConfirmation = false

    init(item: Item, viewModel: ItemsViewModel) {
        self.item = item
        self.viewModel = viewModel
        _selectedCategoryID = State(initialValue: item.categoryID)
        _name = State(initialValue: item.name)
        _isInList = State(initialValue: ItemEditorViewUX.membershipToggleInitialState(isInList: item.list != nil).visibleValue)
    }

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
                    .disabled(ItemEditorViewUX.categoryDisabled(isMutationInFlight: isMutationInFlight))
                    .accessibilityLabel(ItemEditorViewAccessibility.categoryLabel)

                    TextField("Item name", text: $name)
                        .disabled(ItemEditorViewUX.nameDisabled(isMutationInFlight: isMutationInFlight))
                        .accessibilityLabel(ItemEditorViewAccessibility.nameLabel)
                }

                Section("Shopping List") {
                    Toggle(
                        "In shopping list",
                        isOn: Binding(
                            get: { isInList },
                            set: { nextValue in
                                let mutation = ItemEditorViewUX.membershipToggleBeginMutation(
                                    currentVisibleValue: isInList,
                                    requestedValue: nextValue
                                )
                                guard mutation.requestedValue != mutation.visibleValue else {
                                    return
                                }
                                guard !isMutationInFlight else {
                                    return
                                }

                                Task {
                                    let success = await viewModel.setInList(
                                        itemID: item.id,
                                        isInList: mutation.requestedValue
                                    )
                                    isInList = ItemEditorViewUX.membershipToggleResolveMutation(
                                        currentVisibleValue: mutation.visibleValue,
                                        requestedValue: mutation.requestedValue,
                                        success: success
                                    )
                                }
                            }
                        )
                    )
                        .disabled(ItemEditorViewUX.membershipToggleDisabled(isMutationInFlight: isMutationInFlight))
                        .accessibilityLabel(ItemEditorViewAccessibility.membershipToggleLabel)
                }

                Section {
                    Button("Delete Item", role: .destructive) {
                        isShowingDeleteConfirmation = true
                    }
                    .disabled(ItemEditorViewUX.deleteDisabled(isMutationInFlight: isMutationInFlight))
                    .accessibilityLabel(ItemEditorViewAccessibility.deleteButtonLabel)
                }

                if let errorMessage = viewModel.errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .accessibilityLabel(ItemEditorViewAccessibility.errorLabel)
                            .accessibilityValue(errorMessage)
                    }
                }
            }
            .navigationTitle("Edit Item")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                    .disabled(ItemEditorViewUX.cancelDisabled(isMutationInFlight: isMutationInFlight))
                    .accessibilityLabel(ItemEditorViewAccessibility.cancelButtonLabel)
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        Task {
                            let success = await viewModel.updateItem(
                                id: item.id,
                                name: name,
                                categoryID: selectedCategoryID
                            )
                            if success {
                                dismiss()
                            }
                        }
                    }
                    .disabled(
                        ItemEditorViewUX.saveDisabled(
                            isMutationInFlight: isMutationInFlight,
                            baseSaveDisabled: isSaveInputInvalid
                        )
                    )
                    .accessibilityLabel(ItemEditorViewAccessibility.saveButtonLabel)
                }
            }
        }
        .interactiveDismissDisabled(isMutationInFlight)
        .onChange(of: modelIsInList) { _, nextModelIsInList in
            isInList = ItemEditorViewUX.membershipToggleSyncExternalModelChange(
                currentVisibleValue: isInList,
                modelIsInList: nextModelIsInList,
                isMutationInFlight: isMutationInFlight
            )
        }
        .confirmationDialog("Delete this item?", isPresented: $isShowingDeleteConfirmation) {
            Button("Delete", role: .destructive) {
                Task {
                    let success = await viewModel.deleteItem(id: item.id)
                    if success {
                        dismiss()
                    }
                }
            }
            .accessibilityLabel(ItemEditorViewAccessibility.deleteConfirmButtonLabel)

            Button("Cancel", role: .cancel) {}
        } message: {
            Text("This action cannot be undone.")
        }
    }

    private var isMutationInFlight: Bool {
        viewModel.mutatingItemIDs.contains(item.id)
    }

    private var isSaveInputInvalid: Bool {
        name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || selectedCategoryID == nil
    }

    private var modelIsInList: Bool {
        if let updated = viewModel.items.first(where: { $0.id == item.id }) {
            return updated.list != nil
        }
        return item.list != nil
    }
}
