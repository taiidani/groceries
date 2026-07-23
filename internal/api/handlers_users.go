package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) usersListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.ListUsers(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (s *Server) usersGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	user, err := s.db.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "user")
		} else {
			internalError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) usersCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Admin bool   `json:"admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}

	created, err := s.db.CreateUser(r.Context(), models.CreateUserParams{
		Name:  req.Name,
		Admin: req.Admin,
	})
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) usersUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	user, err := s.db.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "user")
		} else {
			internalError(w, err)
		}
		return
	}

	var req struct {
		Name  *string `json:"name"`
		Admin *bool   `json:"admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	updateParams := models.UpdateUserParams{
		ID: user.ID,
	}
	if req.Name != nil {
		updateParams.Name = *req.Name
	}
	if req.Admin != nil {
		updateParams.Admin = *req.Admin
	}

	user, err = s.db.UpdateUser(r.Context(), updateParams)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) usersDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if _, err := s.db.GetUser(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "user")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := s.db.DeleteUser(r.Context(), id); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
