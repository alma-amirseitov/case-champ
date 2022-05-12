package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/redirect/{link:.*}", app.redirect).Methods("GET")

	router.HandleFunc("/admin/redirects", app.authenticate(app.GetRedirects)).Methods("GET")
	router.HandleFunc("/admin/redirects/{id}", app.authenticate(app.GetRedirectById)).Methods("GET")
	router.HandleFunc("admin/redirects", app.CreateRedirect).Methods("POST")
	router.HandleFunc("/admin/redirects/{id:}", app.UpdateRedirect).Methods("PATCH")
	router.HandleFunc("/admin/redirects/{id:}", app.DeleteRedirect).Methods("DELETE")

	router.HandleFunc("/admin", app.registerAdminHandler).Methods("POST")

	router.HandleFunc("/tokens/authentication", app.createAuthenticationTokenHandler).Methods("POST")

	return app.recoverPanic(app.enableCORS(app.rateLimit(router)))
}
