package server

import (
	"fmt"
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

type adminBag struct {
	baseBag
	Users  []models.User
	Groups []models.Group
}

func (s *Server) adminHandler(w http.ResponseWriter, r *http.Request) {
	bag := adminBag{baseBag: s.newBag(r.Context())}

	var err error
	bag.Users, err = s.db.ListUsers(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Groups, err = s.db.ListGroups(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	template := "admin.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	user, err := s.db.GetUser(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	user.Admin = r.FormValue("admin") == "on" || r.FormValue("admin") == "true"
	user.Name = r.FormValue("name")

	user, err = s.db.UpdateUser(r.Context(), models.UpdateUserParams{
		ID:    id,
		Name:  user.Name,
		Admin: user.Admin,
	})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/admin"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) userAddHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.db.CreateUser(r.Context(), models.CreateUserParams{
		Name:  r.FormValue("name"),
		Admin: r.FormValue("admin") == "on" || r.FormValue("admin") == "true",
	})
	if err != nil {
		err = fmt.Errorf("could not add user: %w", err)
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/admin"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = s.db.DeleteUser(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/admin"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) groupUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	group, err := s.db.GetGroup(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	group.Name = r.FormValue("name")

	group, err = s.db.UpdateGroup(r.Context(), models.UpdateGroupParams{
		ID:   id,
		Name: group.Name,
	})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/admin"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) groupAddHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.db.CreateGroup(r.Context(), r.FormValue("name"))
	if err != nil {
		err = fmt.Errorf("could not add group: %w", err)
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/admin"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) groupDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = s.db.DeleteGroup(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/admin"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}
