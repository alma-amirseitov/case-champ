package main

import (
	"case-championship/internal/data"
	"case-championship/internal/validator"
	"net/http"
)

func (app *application) registerAdminHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	admin := &data.Admin{
		Login: input.Login,
	}

	err = admin.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, admin); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user data into the database.
	err = app.models.Admin.Insert(admin)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"admin": admin}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
