package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) groupsListHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := models.LoadGroups(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, groups)
}

func (s *Server) groupsGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	group, err := models.GetGroup(r.Context(), id)
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	group := models.Group{Name: req.Name}
	if err := group.Validate(r.Context()); err != nil {
		badRequest(w, err.Error())
		return
	}

	if err := models.AddGroup(r.Context(), group); err != nil {
		internalError(w, err)
		return
	}

	// Reload to get the assigned ID
	groups, err := models.LoadGroups(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}
	for _, g := range groups {
		if g.Name == req.Name {
			writeJSON(w, http.StatusCreated, g)
			return
		}
	}

	// Fallback: return the input without an ID (should not happen)
	writeJSON(w, http.StatusCreated, group)
}

func (s *Server) groupsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	existing, err := models.GetGroup(r.Context(), id)
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
	if err := existing.Validate(r.Context()); err != nil {
		badRequest(w, err.Error())
		return
	}

	if err := models.EditGroup(r.Context(), existing); err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

func (s *Server) groupsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if _, err := models.GetGroup(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "group")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.DeleteGroup(r.Context(), id); err != nil {
		// DeleteGroup returns a descriptive error when the group is still in use
		conflict(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
