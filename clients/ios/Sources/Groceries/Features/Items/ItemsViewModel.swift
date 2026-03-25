import Foundation
import GroceriesAPI

protocol ItemsAPI: Sendable {
    func listCategories() async throws -> [GroceriesAPI.Category]
    func listItems(inList: Bool?) async throws -> [Item]
    func createItem(categoryID: Int, name: String) async throws -> Item
    func updateItem(id: Int, categoryID: Int, name: String) async throws -> Item
    func deleteItem(id: Int) async throws
    func addItemToList(itemID: Int) async throws -> Item
    func removeItemFromList(itemID: Int) async throws
}

extension GroceriesAPIClient: ItemsAPI {
    func listItems(inList: Bool?) async throws -> [Item] {
        try await listItems(categoryID: nil, inList: inList)
    }

    func addItemToList(itemID: Int) async throws -> Item {
        _ = try await addItemToList(itemID: itemID, quantity: "")
        let items = try await listItems(inList: true)
        guard let item = items.first(where: { $0.id == itemID }) else {
            throw APIError.notFound("Item not found")
        }
        return item
    }
}

@Observable
@MainActor
final class ItemsViewModel {
    private(set) var items: [Item] = []
    private(set) var filteredItems: [Item] = []
    private(set) var categories: [GroceriesAPI.Category] = []
    private(set) var loadErrorMessage: String?
    private(set) var mutationErrorMessage: String?

    private(set) var isLoading = false
    private(set) var isAdding = false
    private(set) var isUpdating = false
    private(set) var mutatingItemIDs: Set<Int> = []

    var isAddCategoryPickerDisabled: Bool { isAdding }
    var isAddNameFieldDisabled: Bool { isAdding }

    var searchText: String = "" {
        didSet { applyFilters() }
    }

    var inListOnly: Bool = false {
        didSet { applyFilters() }
    }

    private let api: any ItemsAPI
    private let notificationCenter: NotificationCenter
    private var activeLoadCount = 0
    private var latestLoadID = 0

    init(api: any ItemsAPI, notificationCenter: NotificationCenter = .default) {
        self.api = api
        self.notificationCenter = notificationCenter
    }

    func load() async {
        await load(force: false)
    }

    func retryLoad() async {
        await load(force: true)
    }

    func refresh() async {
        await load(force: true)
    }

    func addItem(name: String, categoryID: Int?) async -> Bool {
        guard let validated = validatedInput(name: name, categoryID: categoryID) else {
            return false
        }
        guard !isAdding else { return false }

        isAdding = true
        defer { isAdding = false }

        do {
            let item = try await api.createItem(categoryID: validated.categoryID, name: validated.name)
            items.append(item)
            mutationErrorMessage = nil
            applyFilters()
            return true
        } catch {
            mutationErrorMessage = errorDescription(error)
            return false
        }
    }

    func isAddButtonDisabled(name: String, categoryID: Int?) -> Bool {
        guard !isAdding else { return true }
        guard categoryID != nil else { return true }
        return name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    func updateItem(id: Int, name: String, categoryID: Int?) async -> Bool {
        guard let validated = validatedInput(name: name, categoryID: categoryID) else {
            return false
        }
        guard !isUpdating else { return false }
        guard !mutatingItemIDs.contains(id) else { return false }

        isUpdating = true
        mutatingItemIDs.insert(id)
        defer {
            isUpdating = false
            mutatingItemIDs.remove(id)
        }

        do {
            let updated = try await api.updateItem(id: id, categoryID: validated.categoryID, name: validated.name)
            if let index = items.firstIndex(where: { $0.id == id }) {
                items[index] = updated
            }
            mutationErrorMessage = nil
            applyFilters()
            return true
        } catch {
            mutationErrorMessage = errorDescription(error)
            return false
        }
    }

    func deleteItem(id: Int) async -> Bool {
        guard !mutatingItemIDs.contains(id) else { return false }
        mutatingItemIDs.insert(id)
        defer { mutatingItemIDs.remove(id) }

        do {
            try await api.deleteItem(id: id)
            items.removeAll(where: { $0.id == id })
            mutationErrorMessage = nil
            applyFilters()
            return true
        } catch {
            mutationErrorMessage = errorDescription(error)
            return false
        }
    }

    func setInList(itemID: Int, isInList: Bool) async -> Bool {
        guard !mutatingItemIDs.contains(itemID) else { return false }
        guard items.contains(where: { $0.id == itemID }) else { return false }

        mutatingItemIDs.insert(itemID)
        defer { mutatingItemIDs.remove(itemID) }

        do {
            if isInList {
                _ = try await api.addItemToList(itemID: itemID)
            } else {
                try await api.removeItemFromList(itemID: itemID)
            }

            items = try await api.listItems(inList: nil)
            mutationErrorMessage = nil
            applyFilters()

            notificationCenter.post(
                name: AppEvents.MembershipChanged.name,
                object: nil,
                userInfo: [
                    AppEvents.MembershipChanged.itemIDKey: itemID,
                    AppEvents.MembershipChanged.isInListKey: isInList,
                    AppEvents.MembershipChanged.changedAtKey: Date(),
                ]
            )
            return true
        } catch {
            mutationErrorMessage = errorDescription(error)
            return false
        }
    }

    private func load(force: Bool) async {
        guard force || !isLoading else { return }

        latestLoadID += 1
        let loadID = latestLoadID

        activeLoadCount += 1
        isLoading = true
        defer {
            activeLoadCount = max(0, activeLoadCount - 1)
            isLoading = activeLoadCount > 0
        }

        do {
            async let categoriesFetch = api.listCategories()
            async let itemsFetch = api.listItems(inList: nil)
            let (fetchedCategories, fetchedItems) = try await (categoriesFetch, itemsFetch)

            guard loadID == latestLoadID else { return }

            categories = fetchedCategories
            items = fetchedItems
            loadErrorMessage = nil
            applyFilters()
        } catch {
            guard loadID == latestLoadID else { return }
            loadErrorMessage = errorDescription(error)
        }
    }

    private func applyFilters() {
        var next = items

        if inListOnly {
            next = next.filter { $0.list != nil }
        }

        let trimmedSearch = searchText.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmedSearch.isEmpty {
            next = next.filter {
                $0.name.range(of: trimmedSearch, options: .caseInsensitive) != nil
            }
        }

        filteredItems = next
    }

    private func validatedInput(name: String, categoryID: Int?) -> (name: String, categoryID: Int)? {
        let trimmedName = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedName.isEmpty else {
            mutationErrorMessage = "Item name is required."
            return nil
        }
        guard let categoryID else {
            mutationErrorMessage = "Category is required."
            return nil
        }
        return (name: trimmedName, categoryID: categoryID)
    }

    private func errorDescription(_ error: Error) -> String {
        if let apiError = error as? APIError {
            return apiError.errorDescription ?? error.localizedDescription
        }
        return error.localizedDescription
    }
}
