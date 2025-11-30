package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

type adminBag struct {
	baseBag
	Users []models.User
}

func (s *Server) adminHandler(w http.ResponseWriter, r *http.Request) {
	bag := adminBag{baseBag: s.newBag(r.Context())}

	users, err := models.LoadUsers(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	bag.Users = users

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
	err := models.DeleteUser(r.Context(), r.PathValue("id"))
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
