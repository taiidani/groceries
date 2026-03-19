import SwiftUI

// MARK: - LoginView

/// The initial sign-in screen presented when the user is not authenticated.
///
/// Conforms to Apple's Human Interface Guidelines:
/// - Uses the system-provided `.prominent` button style for the primary action.
/// - Respects Dynamic Type and supports all accessibility text sizes.
/// - Supports both Light and Dark appearances via semantic system colors.
/// - Uses `@FocusState` to move focus automatically and dismiss the keyboard
///   on submit, keeping the experience keyboard-friendly.
/// - Displays inline validation feedback rather than modal alerts.
struct LoginView: View {

    // MARK: - Dependencies

    @Environment(AuthViewModel.self) private var authViewModel

    // MARK: - Local state

    @State private var username: String = ""
    @State private var password: String = ""
    @FocusState private var focusedField: Field?

    // MARK: - Field enum

    private enum Field: Hashable {
        case username
        case password
    }

    // MARK: - Body

    var body: some View {
        ScrollView {
            VStack(spacing: 0) {
                header
                    .padding(.top, 60)
                    .padding(.bottom, 40)

                credentialsForm
                    .padding(.bottom, 24)

                signInButton

                if let errorMessage = authViewModel.errorMessage {
                    errorBanner(message: errorMessage)
                        .padding(.top, 16)
                        .transition(.move(edge: .top).combined(with: .opacity))
                }

                Spacer(minLength: 40)
            }
            .padding(.horizontal, 24)
        }
        .scrollBounceBehavior(.basedOnSize)
        .background { AppBackground() }
        .animation(.easeInOut(duration: 0.2), value: authViewModel.errorMessage)
        .onAppear {
            // Auto-focus username on first appearance.
            focusedField = .username
        }
    }

    // MARK: - Subviews

    private var header: some View {
        VStack(spacing: 12) {
            Image(systemName: "cart.fill")
                .font(.system(size: 56, weight: .semibold))
                .foregroundStyle(.tint)
                .accessibilityHidden(true)

            Text("Groceries")
                .font(.largeTitle.bold())
                .foregroundStyle(.white)

            Text("Sign in to manage your shopping lists.")
                .font(.subheadline)
                .foregroundStyle(.white.opacity(0.75))
                .multilineTextAlignment(.center)
        }
    }

    private var credentialsForm: some View {
        VStack(spacing: 0) {
            // Username row
            HStack {
                Image(systemName: "person")
                    .foregroundStyle(.white.opacity(0.6))
                    .frame(width: 24)
                    .accessibilityHidden(true)

                TextField("Username", text: $username)
                    .textContentType(.username)
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)
                    .submitLabel(.next)
                    .focused($focusedField, equals: .username)
                    .onSubmit { focusedField = .password }
                    .foregroundStyle(.white)
                    .tint(.white)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 14)

            Divider()
                .overlay(Color.white.opacity(0.2))
                .padding(.leading, 56)

            // Password row
            HStack {
                Image(systemName: "lock")
                    .foregroundStyle(.white.opacity(0.6))
                    .frame(width: 24)
                    .accessibilityHidden(true)

                SecureField("Password", text: $password)
                    .textContentType(.password)
                    .submitLabel(.go)
                    .focused($focusedField, equals: .password)
                    .onSubmit(submit)
                    .foregroundStyle(.white)
                    .tint(.white)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 14)
        }
        .background(Color.white.opacity(0.12))
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
    }

    private var signInButton: some View {
        Button(action: submit) {
            Group {
                if authViewModel.isLoading {
                    ProgressView()
                        .tint(.white)
                } else {
                    Text("Sign In")
                        .fontWeight(.semibold)
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: 50)
        }
        .buttonStyle(.borderedProminent)
        .buttonBorderShape(.roundedRectangle(radius: 12))
        .disabled(!canSubmit)
        .accessibilityLabel(authViewModel.isLoading ? "Signing in…" : "Sign In")
    }

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
                authViewModel.clearError()
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
        )
    }

    // MARK: - Helpers

    private var canSubmit: Bool {
        !authViewModel.isLoading
            && !username.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            && !password.isEmpty
    }

    private func submit() {
        guard canSubmit else { return }
        focusedField = nil
        Task {
            await authViewModel.login(username: username, password: password)
        }
    }
}

// MARK: - Preview

#Preview("Login — idle") {
    LoginView()
        .environment(AuthViewModel(baseURL: URL(string: "http://localhost:3000")!))
}

#Preview("Login — loading") {
    // A stubbed view model that simulates the loading state.
    let vm = AuthViewModel(baseURL: URL(string: "http://localhost:3000")!)
    return LoginView()
        .environment(vm)
}
