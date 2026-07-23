package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) groupsListHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := s.db.ListGroups(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, groups)
}

func (s *Server) groupsGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	group, err := s.db.GetGroup(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "group")
		} else {
			internalError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, group)
}

func (s *Server) groupsCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		badRequest(w, "invalid request body")
		return
	}

	group := models.Group{Name: req.Name}
	if err := s.db.ValidateGroup(r.Context(), group); err != nil {
		badRequest(w, err.Error())
		return
	}

	group, err = s.db.CreateGroup(r.Context(), req.Name)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, group)
}

func (s *Server) groupsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	existing, err := s.db.GetGroup(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "group")
		} else {
			internalError(w, err)
		}
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	existing.Name = req.Name
	if err := s.db.ValidateGroup(r.Context(), existing); err != nil {
		badRequest(w, err.Error())
		return
	}

	group, err := s.db.UpdateGroup(r.Context(), models.UpdateGroupParams{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, group)
}

func (s *Server) groupsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if _, err := s.db.GetGroup(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "group")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := s.db.DeleteGroup(r.Context(), id); err != nil {
		// DeleteGroup returns a descriptive error when the group is still in use
		conflict(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
