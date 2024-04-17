package main

import (
	"net/http"
	"strconv"
	"strings"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", nil); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminal handles the virtual termainal page for charge card
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

// PaymentSucceeded renders the payment summary page
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.errorLog.Println(err)
	}

	//read posted data
	cardholderName := r.Form.Get("cardholder_name")
	cardholderEmail := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	data := make(map[string]interface{})
	data["cardholderName"] = cardholderName
	data["email"] = cardholderEmail
	data["paymentIntent"] = paymentIntent
	data["paymentMethod"] = paymentMethod
	data["paymentAmount"] = paymentAmount
	data["paymentCurrency"] = paymentCurrency

	if err := app.renderTemplate(w, r, "payment-completed", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// BuyOnce renders the page for buy one widget
func (app *application) BuyOnce(w http.ResponseWriter, r *http.Request) {
	urlparts := strings.Split(r.RequestURI, "/")
	dates_id, _ := strconv.Atoi(urlparts[2])

	date, _ := app.DB.GetDate(dates_id)
	data := make(map[string]interface{})
	data["date"] = date
	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
