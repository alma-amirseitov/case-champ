package main

import (
	"case-championship/internal/data"
	"case-championship/internal/validator"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (app application) redirect(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	key := params["link"]
	var link string
	err := app.cash.View(func(tx *Transaction) error {
		if val, ok := app.cash.data[key]; !ok {
			return errors.New("not found")
		} else {
			link = val
			return nil
		}
	})

	if err == nil {
		http.Redirect(w, r, link, 301)
	} else {
		links, err := app.models.Links.GetByLink(key)

		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		app.cash.Update(func(tx *Transaction) error {
			tx.Set(key, links.ActiveLink)
			return nil
		})

		http.Redirect(w, r, links.ActiveLink, 301)
	}
}
func (app *application) GetRedirectById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	links, err := app.models.Links.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"link": links}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app application) GetRedirects(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	links, err := app.models.Links.GetAll(input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"links": links}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) CreateRedirect(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ActiveLink  string `json:"active_link"`
		HistoryLink string `json:"history_link"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	links := &data.Links{
		ActiveLink:  input.ActiveLink,
		HistoryLink: input.HistoryLink,
	}

	err = app.models.Links.Insert(links)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("link", fmt.Sprintf("api/links/%d", links.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"links": links}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) UpdateRedirect(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	links, err := app.models.Links.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		ActiveLink *string `json:"active_link"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.ActiveLink != nil {
		links.HistoryLink = links.ActiveLink
		links.ActiveLink = *input.ActiveLink
	}

	err = app.models.Links.Update(links)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"links": links}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) DeleteRedirect(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Links.Delete(int64(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "link successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
