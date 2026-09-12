import SwiftUI

// MARK: - LoginView

/// The initial sign-in screen presented when the user is not authenticated.
///
/// Sign-in is delegated entirely to Authelia via a system web view
/// (`ASWebAuthenticationSession`) — there is no local credentials form.
///
/// Conforms to Apple's Human Interface Guidelines:
/// - Uses the system-provided `.prominent` button style for the primary action.
/// - Respects Dynamic Type and supports all accessibility text sizes.
/// - Supports both Light and Dark appearances via semantic system colors.
/// - Displays inline validation feedback rather than modal alerts.
struct LoginView: View {

    // MARK: - Dependencies

    @Environment(AuthViewModel.self) private var authViewModel

    // MARK: - Body

    var body: some View {
        ScrollView {
            VStack(spacing: 0) {
                header
                    .padding(.top, 60)
                    .padding(.bottom, 40)

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

            Text("Sign in with your Authelia account to manage your shopping lists.")
                .font(.subheadline)
                .foregroundStyle(.white.opacity(0.75))
                .multilineTextAlignment(.center)
        }
    }

    private var signInButton: some View {
        Button(action: submit) {
            Group {
                if authViewModel.isLoading {
                    ProgressView()
                        .tint(.white)
                } else {
                    Text("Sign in with Authelia")
                        .fontWeight(.semibold)
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: 50)
        }
        .buttonStyle(.borderedProminent)
        .buttonBorderShape(.roundedRectangle(radius: 12))
        .disabled(authViewModel.isLoading)
        .accessibilityLabel(authViewModel.isLoading ? "Signing in…" : "Sign in with Authelia")
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

    private func submit() {
        Task {
            await authViewModel.login()
        }
    }
}

// MARK: - Preview

#Preview("Login — idle") {
    LoginView()
        .environment(AuthViewModel(baseURL: URL(string: "http://localhost:3000")!))
}

#Preview("Login — loading") {
    let vm = AuthViewModel(baseURL: URL(string: "http://localhost:3000")!)
    return LoginView()
        .environment(vm)
}
