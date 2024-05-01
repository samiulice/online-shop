package main

import (
	"net/http"
	"online_store/internal/cards"
	"online_store/internal/models"
	"strconv"
	"strings"
	"time"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", nil); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminal handles the virtual termainal page for charge card
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminalPaymentSucceeded handles post request of the payment succeeded for virtual terminal
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//save transaction info to the database
	txn := models.Transaction{
		Amount:              txnData.Amount,
		Currency:            txnData.Currency,
		PaymentIntent:       txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		LastFourDigits:      txnData.LastFourDigits,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2, // cleared payment in this case
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//Saving receipt info to the session
	app.Session.Put(r.Context(), "receipt", txnData)

	//redirecting to a new page so that user can't accidently resubmit the form
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

// VirtualTerminalReceipt renders the payment summary for any transaction for the virtual terminal
func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {

	//Retriving receipt info from the session
	txnData := app.Session.Get(r.Context(), "receipt").(models.TransactionData)
	data := make(map[string]interface{})
	data["txnData"] = txnData

	//Removing receipt info from the session
	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// BuyOnce renders the page for buy single package dates
func (app *application) BuyOnce(w http.ResponseWriter, r *http.Request) {
	urlparts := strings.Split(r.RequestURI, "/")
	dates_id, _ := strconv.Atoi(urlparts[2])

	date, _ := app.DB.GetDate(dates_id)
	data := make(map[string]interface{})
	data["dates"] = date
	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js-one-off"); err != nil {
		app.errorLog.Println(err)
	}
}

// PaymentSucceeded handles post request of the payment succeeded
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.errorLog.Println(err)
	}

	//read posted data
	datesID := r.Form.Get("package_id")

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//save customer info to the database
	c := models.Customer{
		FirstName: txnData.FirstName,
		LastName:  txnData.LastName,
		Email:     txnData.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	customerID, err := app.SaveCustomer(c)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//save transaction info to the database
	txn := models.Transaction{
		Amount:              txnData.Amount,
		Currency:            txnData.Currency,
		PaymentIntent:       txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		LastFourDigits:      txnData.LastFourDigits,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2, // cleared payment in this case
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	transactionID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	id := 0
	if !strings.Contains(r.Referer(), "virtual-terminal") {
		//no dates id exist for virtual terminal
		//in that case, dates id needs to be updated later
		//save order info to the database
		id, err = strconv.Atoi(datesID)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}
	order := models.Order{
		DatesID:       id,
		TransactionID: transactionID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        txn.Amount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//Saving receipt info to the session
	app.Session.Put(r.Context(), "receipt", txnData)

	//redirecting to a new page so that user can't accidently resubmit the form
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

// Receipt renders the payment summary for any transaction
func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {

	//Retriving receipt info from the session
	txnData := app.Session.Get(r.Context(), "receipt").(models.TransactionData)
	data := make(map[string]interface{})
	data["txnData"] = txnData

	//Removing receipt info from the session
	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	dates, err := app.DB.GetDate(2) //ID = 2 for Bronze Plan
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	data := map[string]interface{}{
		"dates": dates,
	}
	err = app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	})
	// err = app.renderTemplate(w, r, "bronze-plan", &templateData{
	// 	Data: data,
	// }, "stripe-js-recurring")

	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

// BronzePlanReceipt renders the payment summary for Bronze plan
func (app *application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "bronze-receipt", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// Signin renders the Signin page for the app user
func (app *application) Signin(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "signin", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// PostSignin handles post signin request
func (app *application) PostSignin(w http.ResponseWriter, r *http.Request) {
	app.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	user_id := r.Form.Get("user_id")
	app.Session.Put(r.Context(), "user_id", user_id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
// SignOut helps to sign out an user
func (app *application)  SignOut(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

// PageNotFound renders 404 page not found
func (app *application) PageNotFound(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "page-not-found", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// .........Helper functions for the handlers............//
// SaveCustomer takes customer info as parameters, saves it to the database and returns its id
func (app *application) SaveCustomer(c models.Customer) (int, error) {
	var id int

	id, err := app.DB.InsertCustomer(c)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SaveTransaction takes transaction info as parameters, saves it to the database and returns its id
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	var id int

	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SaveOrder takes SaveOrder info as parameters, saves it to the database and returns its id
func (app *application) SaveOrder(order models.Order) (int, error) {
	var id int

	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetTransactionData gets transaction data from post and stripe
func (app *application) GetTransactionData(r *http.Request) (models.TransactionData, error) {
	var txnData models.TransactionData

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	//read posted data
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	cardHolderEmail := r.Form.Get("cardholder_email")
	cardHolderName := r.Form.Get("cardholder_name")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentCurrency := r.Form.Get("payment_currency")
	paymentAmount := r.Form.Get("payment_amount")

	amount, err := strconv.Atoi(paymentAmount)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}
	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.secret,
	}

	pi, err := card.RetrivePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear
	bankReturnCode := pi.Charges.Data[0].ID

	//Fill txnData
	txnData.FirstName = firstName
	txnData.LastName = lastName
	txnData.Email = cardHolderEmail
	txnData.NameOnCard = cardHolderName
	txnData.Amount = amount
	txnData.Currency = paymentCurrency
	txnData.PaymentAmount = paymentAmount
	txnData.PaymentIntent = paymentIntent
	txnData.PaymentMethod = paymentMethod
	txnData.LastFourDigits = lastFour
	txnData.BankReturnCode = bankReturnCode
	txnData.ExpiryMonth = int(expiryMonth)
	txnData.ExpiryYear = int(expiryYear)

	return txnData, err
}
