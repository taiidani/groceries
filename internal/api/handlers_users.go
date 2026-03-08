package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) usersListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := models.LoadUsers(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (s *Server) usersGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	user, err := models.GetUser(r.Context(), id)
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

	newUser := models.User{
		Name:  req.Name,
		Admin: req.Admin,
	}

	if err := models.AddUser(r.Context(), newUser); err != nil {
		internalError(w, err)
		return
	}

	created, err := models.GetUserByCredentials(r.Context(), req.Name, "")
	if err != nil {
		// AddUser succeeded; we just can't fetch by credentials without a password.
		// Return a minimal representation instead.
		writeJSON(w, http.StatusCreated, newUser)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) usersUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	user, err := models.GetUser(r.Context(), id)
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

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Admin != nil {
		user.Admin = *req.Admin
	}

	if err := models.EditUser(r.Context(), user); err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) usersDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if _, err := models.GetUser(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "user")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.DeleteUser(r.Context(), id); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
