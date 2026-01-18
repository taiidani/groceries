package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

const sseEventRecipe = "recipe"

func (s *Server) recipesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Recipes []models.Recipe
	}

	bag := data{baseBag: s.newBag(r.Context())}

	var err error
	bag.Recipes, err = models.LoadRecipes(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	template := "recipes.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) recipeHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Recipe models.Recipe
		Items  []models.Item
	}

	bag := data{baseBag: s.newBag(r.Context())}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	bag.Recipe, err = models.GetRecipe(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Load all items for the dropdown
	bag.Items, err = models.LoadItems(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	template := "recipe.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) recipeAddHandler(w http.ResponseWriter, r *http.Request) {
	newRecipe := models.Recipe{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	id, err := models.AddRecipe(r.Context(), newRecipe)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventRecipe, nil)

	http.Redirect(w, r, "/recipe/"+strconv.Itoa(id), http.StatusFound)
}

func (s *Server) recipeEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	recipe := models.Recipe{
		ID:          id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	err = models.EditRecipe(r.Context(), recipe)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventRecipe, nil)

	http.Redirect(w, r, "/recipe/"+strconv.Itoa(id), http.StatusFound)
}

func (s *Server) recipeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	err = models.DeleteRecipe(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventRecipe, nil)

	http.Redirect(w, r, "/recipes", http.StatusFound)
}

func (s *Server) recipeAddItemHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	itemID, err := strconv.Atoi(r.FormValue("itemID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	quantity := r.FormValue("quantity")

	err = models.AddRecipeItem(r.Context(), recipeID, itemID, quantity)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventRecipe, nil)

	http.Redirect(w, r, "/recipe/"+strconv.Itoa(recipeID), http.StatusFound)
}

func (s *Server) recipeRemoveItemHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	itemID, err := strconv.Atoi(r.PathValue("itemID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	err = models.RemoveRecipeItem(r.Context(), recipeID, itemID)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventRecipe, nil)

	http.Redirect(w, r, "/recipe/"+strconv.Itoa(recipeID), http.StatusFound)
}

func (s *Server) recipeAddToListHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	itemID, err := strconv.Atoi(r.FormValue("itemID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	quantity := r.FormValue("quantity")

	// Try to add item - ListAddItem will fail if already exists due to unique constraint
	err = models.ListAddItem(r.Context(), itemID, quantity)
	if err != nil {
		// Ignore duplicate errors, just redirect
		if err.Error() != "" {
			// Check if it's a duplicate key error - if so, ignore it
			// Otherwise, return the error
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	http.Redirect(w, r, "/recipe/"+strconv.Itoa(recipeID), http.StatusFound)
}

func (s *Server) recipeAddAllToListHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	err = models.AddRecipeToList(r.Context(), recipeID)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	http.Redirect(w, r, "/recipe/"+strconv.Itoa(recipeID), http.StatusFound)
}
