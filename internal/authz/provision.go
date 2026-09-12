package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/taiidani/groceries/internal/db/models"
)

// SyncUserFromOIDC finds or creates the local user record for the given
// OIDC-provisioned username, and syncs the admin flag from the caller's
// group memberships (a user is considered an admin iff they're in the
// "admins" group). Used by every login path backed by Authelia: the web
// OIDC callback, the web dev-login bypass, and the API's token exchange.
func SyncUserFromOIDC(ctx context.Context, db *models.Queries, username string, groups []string) (models.User, error) {
	isAdmin := slices.Contains(groups, "admins")

	user, err := db.GetUserByName(ctx, username)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		user, err = db.CreateUser(ctx, models.CreateUserParams{Name: username, Admin: isAdmin})
		if err != nil {
			return models.User{}, fmt.Errorf("could not create user: %w", err)
		}
		slog.InfoContext(ctx, "Provisioned new user from OIDC login", "name", user.Name, "admin", user.Admin)
		return user, nil
	case err != nil:
		return models.User{}, fmt.Errorf("could not look up user: %w", err)
	case user.Admin != isAdmin:
		// Authelia's `groups` claim is the source of truth for admin status;
		// re-sync it on every login.
		user, err = db.UpdateUser(ctx, models.UpdateUserParams{ID: user.ID, Name: user.Name, Admin: isAdmin})
		if err != nil {
			return models.User{}, fmt.Errorf("could not sync admin status: %w", err)
		}
		return user, nil
	default:
		return user, nil
	}
}
