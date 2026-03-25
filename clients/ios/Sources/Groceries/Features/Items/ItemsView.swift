import SwiftUI
import GroceriesAPI

struct ItemsView: View {
    let apiClient: GroceriesAPIClient

    var body: some View {
        NavigationStack {
            Text("Items")
        }
    }
}
