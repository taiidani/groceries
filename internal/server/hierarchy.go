package server

import (
	"context"

	"github.com/taiidani/groceries/internal/service"
)

// loadStoreHierarchy delegates to the service layer, adapting the server's
// filter input shape onto the service's.
func (s *Server) loadStoreHierarchy(ctx context.Context, input storeHierarchyInput) ([]storeWithCategories, error) {
	return s.svc.LoadStoreHierarchy(ctx, service.HierarchyInput{
		ExcludeEmptyGroupings: input.ExcludeEmptyGroupings,
		ExcludeDoneItems:      input.ExcludeDoneItems,
		OnlyListItems:         input.OnlyListItems,
	})
}

type storeHierarchyInput struct {
	ExcludeEmptyGroupings bool
	ExcludeDoneItems      bool
	OnlyListItems         bool
}
