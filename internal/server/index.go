package server

import (
	"errors"
	"net/http"

	"github.com/taiidani/groceries/internal/data"
	"github.com/taiidani/groceries/internal/models"
)

type indexBag struct {
	baseBag
	List models.List
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r)}

	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}
	bag.List = list

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}
