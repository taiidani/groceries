package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/taiidani/groceries/internal/models"
)

type categoryWithItems struct {
	models.Category
	Items []models.Item
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	type indexBag struct {
		baseBag
	}

	bag := indexBag{baseBag: s.newBag(r.Context())}

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) indexBagHandler(w http.ResponseWriter, r *http.Request) {
	html, err := s.indexBag(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, html)
}

func (s *Server) indexBag(ctx context.Context) (string, error) {
	type indexBagBag struct {
		baseBag
		Items    []models.Item
		BagItems []models.Item
	}

	bag := indexBagBag{baseBag: s.newBag(ctx)}

	items, err := models.LoadItems(ctx)
	if err != nil {
		return "", err
	}
	bag.Items = items

	bagItems, err := models.LoadBag(ctx)
	if err != nil {
		return "", err
	}
	bag.BagItems = bagItems

	return returnHtml("index_bag.gohtml", bag), nil
}

func (s *Server) indexListHandler(w http.ResponseWriter, r *http.Request) {
	html, err := s.indexList(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, html)
}

func (s *Server) indexList(ctx context.Context) (string, error) {
	type indexListBag struct {
		baseBag
		ListCategories []categoryWithItems
	}

	bag := indexListBag{baseBag: s.newBag(ctx)}

	categories, err := models.LoadCategories(ctx)
	if err != nil {
		return "", err
	}

	listItems, err := models.LoadList(ctx)
	if err != nil {
		return "", err
	}

	for _, cat := range categories {
		addList := []models.Item{}
		for _, item := range listItems {
			if item.CategoryID != cat.ID {
				continue
			} else if !item.List.Done {
				addList = append(addList, item)
			}
		}

		if len(addList) > 0 {
			bag.ListCategories = append(bag.ListCategories, categoryWithItems{
				Category: models.Category{
					ID:          cat.ID,
					Description: cat.Description,
					Name:        cat.Name,
				},
				Items: addList,
			})
		}
	}

	return returnHtml("index_list.gohtml", bag), nil
}

func (s *Server) indexCartHandler(w http.ResponseWriter, r *http.Request) {
	html, err := s.indexCart(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, html)
}

func (s *Server) indexCart(ctx context.Context) (string, error) {
	type indexCartBag struct {
		baseBag
		Total          int
		TotalDone      int
		Categories     []models.Category
		BagItems       []models.Item
		DoneCategories []categoryWithItems
	}

	bag := indexCartBag{baseBag: s.newBag(ctx)}

	categories, err := models.LoadCategories(ctx)
	if err != nil {
		return "", err
	}
	bag.Categories = categories

	bagItems, err := models.LoadBag(ctx)
	if err != nil {
		return "", err
	}
	bag.BagItems = bagItems

	listItems, err := models.LoadList(ctx)
	if err != nil {
		return "", err
	}

	for _, cat := range categories {
		addDone := []models.Item{}
		for _, item := range listItems {
			if item.CategoryID != cat.ID {
				continue
			} else if item.List.Done {
				bag.Total++
				bag.TotalDone++
				addDone = append(addDone, item)
			} else {
				bag.Total++
			}
		}

		if len(addDone) > 0 {
			bag.DoneCategories = append(bag.DoneCategories, categoryWithItems{
				Category: models.Category{
					ID:          cat.ID,
					Description: cat.Description,
					Name:        cat.Name,
				},
				Items: addDone,
			})
		}
	}

	return returnHtml("index_cart.gohtml", bag), nil
}
