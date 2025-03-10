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
	mux.Get("/setup-new-password", app.SetupNewUserPassword)

	//404 not found route
	mux.NotFound(app.PageNotFound)
	mux.Get("/test", app.Test)

	//Public file server
	
	publicFileServer := http.FileServer(http.Dir("./public/admin"))
	mux.Handle("/public/admin/*", http.StripPrefix("/public/admin", publicFileServer))

	//secure routes
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/virtual-terminal", app.VirtualTerminal)
		mux.Get("/dashboard", app.AdminDashboard)

		//routes for general >> Admin
		mux.Get("/general/profile/view", app.AdminViewProfile)
		mux.Get("/general/user/add", app.AdminAddUser)

		//routes for analytics >> Service Provider
		mux.Get("/analytics/employees/active", app.AdminViewEmployee)
		mux.Get("/analytics/employees/ex", app.AdminViewEmployee)
		mux.Get("/analytics/employees/suspended", app.AdminViewEmployee)
		mux.Get("/analytics/employees/resigned", app.AdminViewEmployee)
		mux.Get("/analytics/employees/all", app.AdminViewEmployee)
		mux.Get("/analytics/employees/profile/view/{id}", app.AdminViewEmployee)

		//routes for business analytics >> order
		mux.Get("/analytics/order/view/{id}", app.AdminOrderHistoy)

		//routes for business analytics >> transaction-history
		mux.Get("/analytics/transaction/view/{id}", app.AdminViewTransaction)

		//routes for customer management >> view >> profile
		mux.Get("/customer/profile/view/{id}", app.AdminViewCustomerProfile)

		//404 not found route
		mux.NotFound(app.PageNotFound)


		//Admin file server
		publicFileServer := http.FileServer(http.Dir("./"))
		mux.Handle("/*", http.StripPrefix("/", publicFileServer))
	})

	// Secure routes for employees
	mux.Route("/employee", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/dashboard", app.EmployeeDashboard)

		// 404 not found route for employee section
		mux.NotFound(app.PageNotFound)

		// Employee file server
		employeeFileServer := http.FileServer(http.Dir("./public"))
		mux.Handle("/*", http.StripPrefix("/", employeeFileServer))
	})

	return mux
}
