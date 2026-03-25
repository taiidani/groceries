package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/taiidani/groceries/internal/models"
)

func TestAccountHandlerRendersUsername(t *testing.T) {
	origDevMode := DevMode
	DevMode = false
	t.Cleanup(func() { DevMode = origDevMode })

	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/account", nil)
	ctx := context.WithValue(req.Context(), userKey, &models.User{Name: "alice"})
	rec := httptest.NewRecorder()

	s.accountHandler(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "alice") {
		t.Fatalf("expected response body to include username, got %q", rec.Body.String())
	}
}

func TestIndexListTemplateDoesNotContainLargeHeading(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("templates/index_list.gohtml")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	if strings.Contains(string(b), "Grocery List") {
		t.Fatalf("expected shopping list heading to be removed")
	}
}

func TestFooterTemplateContainsBottomNavigationLinks(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("templates/footer.gohtml")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	body := string(b)
	if !strings.Contains(body, "href=\"/\"") {
		t.Fatalf("expected footer to include list navigation link")
	}

	if !strings.Contains(body, "href=\"/account\"") {
		t.Fatalf("expected footer to include account navigation link")
	}
}
