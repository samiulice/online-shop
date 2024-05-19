package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Post("/api/payment-intent", app.GetPaymentIntent)
	mux.Post("/api/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribe)
	mux.Post("/api/authenticate", app.CreateAuthToken)
	mux.Post("/api/is-authenticated", app.CheckAuthenticated)
	mux.Post("/api/forgot-password", app.ForgotPassword)
	mux.Post("/api/reset-password", app.ResetPassword)
	mux.Post("/api/setup-new-password", app.SetupNewUserPassword)

	//Secure routes
	mux.Route("/api/admin", func(mux chi.Router) {
		mux.Use(app.Auth)

		mux.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceeded)

		//general
		// mux.Post("/general/employees/add/send-verification-code", app.SendVerificationCode)
		mux.Post("/general/user/add", app.AdminAddUser)
		// mux.Post("/general/employees/edit/{id}", app.UpdateEmployeeAccount) //suspend all the authority an account temporarily
		mux.Post("/general/employees/activate/{id}", app.ManageEmployeeAccount) //active all the authority a suspened account 
		
		mux.Post("/general/employees/suspend/{id}", app.ManageEmployeeAccount) //suspend all the authority an account temporarily
		mux.Post("/general/employees/revoke/{id}", app.ManageEmployeeAccount) //revoke = suspend all the authority an account permanently
		mux.Post("/general/employees/rejoin/{id}", app.ManageEmployeeAccount) //undo deleted account-rejoin into the job
		
		//Order
		mux.Post("/analytics/order/view/{type}", app.GetOrdersHistoy)
		mux.Post("/analytics/order/refund/{type}", app.RefundCharge)
		mux.Post("/analytics/subscription/cancel/{type}", app.CancelSubscription)
		
		//transaction
		mux.Post("/analytics/transaction/view/{type}", app.GetTransactionHistory)
		//employee
		mux.Post("/analytics/employee/{type}", app.GetEmployees)
		
		//cusotmer
		mux.Post("/customer/profile/view/{type}", app.AdminCustomerProfile)
	})
	return mux
}
