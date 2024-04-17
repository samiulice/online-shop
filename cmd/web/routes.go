package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Get("/", app.Home)
	mux.Get("/buy-dates/{id}", app.BuyOnce)

	mux.Get("/virtual-terminal", app.VirtualTerminal)


	mux.Post("/payment-succeeded", app.PaymentSucceeded)

	//Public file server
	publicFileServer := http.FileServer(http.Dir("./public/assets"))
	mux.Handle("/public/assets/*", http.StripPrefix("/public/assets", publicFileServer))
	return mux
}
