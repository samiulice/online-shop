package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(SessionLoad)

	mux.Get("/", app.Home)

	mux.Get("/buy-dates/{id}", app.BuyOnce)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)
	mux.Get("/receipt", app.Receipt)

	mux.Get("/plans/bronze", app.BronzePlan)
	mux.Get("/receipt/bronze", app.BronzePlanReceipt)

	//Auhtentication
	mux.Get("/signin", app.Signin)
	mux.Post("/signin", app.PostSignin)
	mux.Get("/signout", app.SignOut)

	//Reset Password
	mux.Get("/forgot-password", app.ForgotPassword)
	mux.Get("/reset-password", app.ResetPassword)

	//404 not found route
	mux.NotFound(app.PageNotFound)
	mux.Get("/test", app.Test)

	//Public file server
	publicFileServer := http.FileServer(http.Dir("./public/assets"))
	mux.Handle("/public/assets/*", http.StripPrefix("/public/assets", publicFileServer))

	//secure routes
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/virtual-terminal", app.VirtualTerminal)
		mux.Get("/dashboard", app.AdminDashboard)

		//routes for business analytics >> order
		mux.Get("/analytics/order/view/{id}", app.AdminOrderHistoy)

		//routes for business analytics >> transaction-history
		mux.Get("/analytics/transaction/view/{id}", app.AdminViewTransaction)

		//routes for customer management >> view >> profile
		mux.Get("/customer/profile/view/{id}", app.AdminViewCustomerProfile)

		//Admin file server
		publicFileServer := http.FileServer(http.Dir("./"))
		mux.Handle("/*", http.StripPrefix("/", publicFileServer))
	})

	return mux
}
