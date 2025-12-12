package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

type adminBag struct {
	baseBag
	Users  []models.User
	Groups []models.Group
}

func (s *Server) adminHandler(w http.ResponseWriter, r *http.Request) {
	bag := adminBag{baseBag: s.newBag(r.Context())}

	var err error
	bag.Users, err = models.LoadUsers(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Groups, err = models.LoadGroups(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	template := "admin.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	user, err := models.GetUser(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	user.Admin = r.FormValue("admin") == "on" || r.FormValue("admin") == "true"
	user.Name = r.FormValue("name")

	err = models.EditUser(r.Context(), user)
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
	newUser := models.User{
		Name:  r.FormValue("name"),
		Admin: r.FormValue("admin") == "on" || r.FormValue("admin") == "true",
	}

	err := models.AddUser(r.Context(), newUser)
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = models.DeleteUser(r.Context(), id)
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
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	group, err := models.GetGroup(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	group.Name = r.FormValue("name")

	err = models.EditGroup(r.Context(), group)
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
	newGroup := models.Group{
		Name: r.FormValue("name"),
	}

	err := models.AddGroup(r.Context(), newGroup)
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = models.DeleteGroup(r.Context(), id)
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
