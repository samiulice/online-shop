package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"online_store/internal/cards"
	"online_store/internal/encryption"
	"online_store/internal/models"
	"online_store/internal/urlsigner"
	"path"
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
	if err := app.renderTemplate(w, r, "admin-virtual-terminal", &templateData{}); err != nil {
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

// BuyOnce renders the page for buy a pair of boots
func (app *application) BuyOnce(w http.ResponseWriter, r *http.Request) {
	urlparts := strings.Split(r.RequestURI, "/")
	dates_id, _ := strconv.Atoi(urlparts[2])

	date, _ := app.DB.GetDate(dates_id)
	data := make(map[string]interface{})
	data["product"] = date
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
		UserName:  strings.Split(txnData.Email, "@")[0],
		FirstName: txnData.FirstName,
		LastName:  txnData.LastName,
		Email:     txnData.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	customerID, err := app.SaveCustomer(c)
	if err != nil {
		app.errorLog.Println("ErrorInsertOrder: ",err)
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

	//Get product info
	p, err := app.DB.GetDate(order.DatesID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	//call invoice microservice to generate invoice template and send it to the customer email address
	var product = models.InvoiceProduct{
		ID:       order.DatesID,
		Name:     p.Name,
		Quantity: order.Quantity,
		Amount:   order.Amount,
	}
	var items = []models.InvoiceProduct{product}
	var inv = models.Invoice{
		ID:        order.ID,
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Email:     c.Email,
		CreatedAt: time.Now(),
		Items:     items,
	}
	err = app.callInvoiceMicro(inv)
	if err != nil {
		app.errorLog.Println(err)
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

// BronzePlan renders the page for buy a pair of boots each month
func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	dates, err := app.DB.GetDate(2) //ID = 2 for Bronze Plan
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	data := map[string]interface{}{
		"product": dates,
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
	if app.Session.Exists(r.Context(), "user_id") {
		account_type := app.Session.Get(r.Context(),"account_type").(string)
		if account_type == "employees" {
			http.Redirect(w, r, fmt.Sprintf("public/%s/dashboard", account_type[:len(account_type)-1]), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/%s/dashboard", account_type[:len(account_type)-1]), http.StatusSeeOther)
		}
	} else {
		err := app.renderTemplate(w, r, "signin", &templateData{})
		if err != nil {
			app.errorLog.Println(err)
		}
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
	account_type := r.Form.Get("account_type")
	user, err := app.DB.GetUserDetails(user_id, "id", account_type)
	user.AccountType = account_type
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.Session.Put(r.Context(), "user_id", user_id)
	app.Session.Put(r.Context(), "account_type", account_type)
	app.Session.Put(r.Context(), "user", user)
	if len(account_type) > 0 {
		if account_type == "employees" {
			http.Redirect(w, r, fmt.Sprintf("public/%s/dashboard", account_type[:len(account_type)-1]), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/%s/dashboard", account_type[:len(account_type)-1]), http.StatusSeeOther)
		}
	}	
}

// SignOut helps to sign out an user
func (app *application) SignOut(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

// ForgotPassword renders forget password page for the user
func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// ResetPassword renders reset password page from signed url
func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	//verify that url was signed
	url := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.config.frontend, url)

	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretKey),
	}

	//Verify and check Token expiry
	valid := signer.VerifyToken(testURL)

	data := make(map[string]interface{})
	if !valid {
		data["msg"] = "tempered or broken"
		if err := app.renderTemplate(w, r, "password-reset-link-invalid", &templateData{Data: data}); err != nil {
			app.errorLog.Println(err)
			return
		}
		return
	}
	expired := signer.Expired(testURL, 60)
	if expired {
		data["msg"] = "expired"
		if err := app.renderTemplate(w, r, "password-reset-link-invalid", &templateData{Data: data}); err != nil {
			app.errorLog.Println(err)
			return
		}
		return
	}
	email := r.URL.Query().Get("email")
	userID := r.URL.Query().Get("user_id")
	userType := r.URL.Query().Get("user")

	//encrypt email and userID
	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretKey),
	}

	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		app.errorLog.Println("falied to encrypt email:\t", err)
		return
	}
	encryptedUserID, err := encryptor.Encrypt(userID)
	if err != nil {
		app.errorLog.Println("falied to encrypt userID:\t", err)
		return
	}

	data["email"] = encryptedEmail
	data["user_id"] = encryptedUserID
	data["user"] = userType

	if err := app.renderTemplate(w, r, "reset-password", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
	}

}

// ResetPassword renders reset password page from signed url
func (app *application) SetupNewUserPassword(w http.ResponseWriter, r *http.Request) {
	//verify that url was signed
	url := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.config.frontend, url)

	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretKey),
	}

	//Verify and check Token expiry
	valid := signer.VerifyToken(testURL)

	data := make(map[string]interface{})
	if !valid {
		data["msg"] = "tempered or broken"
		if err := app.renderTemplate(w, r, "password-reset-link-invalid", &templateData{Data: data}); err != nil {
			app.errorLog.Println(err)
			return
		}
		return
	}
	expired := signer.Expired(testURL, 60)
	if expired {
		data["msg"] = "expired"
		if err := app.renderTemplate(w, r, "password-reset-link-invalid", &templateData{Data: data}); err != nil {
			app.errorLog.Println(err)
			return
		}
		return
	}
	email := r.URL.Query().Get("email")
	userID := r.URL.Query().Get("user_id")
	userType := r.URL.Query().Get("user")

	//encrypt email and userID
	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretKey),
	}

	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		app.errorLog.Println("falied to encrypt email:\t", err)
		return
	}
	encryptedUserID, err := encryptor.Encrypt(userID)
	if err != nil {
		app.errorLog.Println("falied to encrypt userID:\t", err)
		return
	}

	data["email"] = encryptedEmail
	data["user_id"] = encryptedUserID
	data["user"] = userType

	if err := app.renderTemplate(w, r, "setup-new-password", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
	}

}

// PageNotFound renders 404 page not found
func (app *application) PageNotFound(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "page-not-found", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// Test renders pages for testing purposes
func (app *application) Test(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "test-html", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// .........Handler function for Admin Panel............//
// AdminDashboard renders admin dashboard
func (app *application) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	user := app.Session.Get(r.Context(), "user").(models.User)
	if err := app.renderTemplate(w, r, "admin-dashboard", &templateData{User: user}); err != nil {
		app.errorLog.Println(err)
	}
}

// AdminViewProfile renders admin profile page
func (app *application) AdminViewProfile(w http.ResponseWriter, r *http.Request) {
	user := app.Session.Get(r.Context(), "user").(models.User)
	if err := app.renderTemplate(w, r, "admin-view-profile", &templateData{User: user}); err != nil {
		app.errorLog.Println(err)
	}
}

// AdminAddEmployee renders add admin page
func (app *application) AdminAddUser(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query().Get("user")
	var tmpl = "admin-add-"
	if userType == "employee" {
		tmpl += userType
	} else {
		tmpl += "user"
	}
	user := app.Session.Get(r.Context(), "user").(models.User)
	if err := app.renderTemplate(w, r, tmpl, &templateData{User: user}); err != nil {
		app.errorLog.Println(err)
	}
}

// AdminEmployeeList renders employee list
func (app *application) AdminViewEmployee(w http.ResponseWriter, r *http.Request) {
	t := path.Base(r.URL.Path)

	data := make(map[string]interface{})
	data["employee-list-type"] = t
	user := app.Session.Get(r.Context(), "user").(models.User)

	tmpl := ""
	if t == "active" || t == "ex" || t == "suspended" || t == "resigned" || t == "all" {
		tmpl = "admin-employee-list"
	} else if _, err := strconv.Atoi(t); err == nil {
		tmpl = "admin-employee-details"
	} else {
		app.errorLog.Println("Invlaid customer id: ", t)
		http.Redirect(w, r, "/page-not-found", http.StatusNotFound)
		return
	}

	err := app.renderTemplate(w, r, tmpl, &templateData{
		User: user,
		Data: data,
	})
	if err != nil {
		app.errorLog.Println(err)
	}
}

// AdminSalesHistoy renders various sales history
func (app *application) AdminOrderHistoy(w http.ResponseWriter, r *http.Request) {

	t := path.Base(r.URL.Path)
	data := make(map[string]interface{})
	data["history-type"] = t
	user := app.Session.Get(r.Context(), "user").(models.User)
	tmpl := ""
	if t == "all" || t == "completed" || t == "processing" || t == "refunded" || t == "cancelled" || t == "one-off" || t == "subscriptions" {
		tmpl = "admin-orders-list"
	} else if _, err := strconv.Atoi(t); err == nil {
		tmpl = "admin-order-details"
	} else {
		app.errorLog.Println("Invlaid customer id: ", t)
		http.Redirect(w, r, "/page-not-found", http.StatusNotFound)
		return
	}

	err := app.renderTemplate(w, r, tmpl, &templateData{
		User: user,
		Data: data,
	})
	if err != nil {
		app.errorLog.Println(err)
	}
}

// .........Handler function for Employee Panel............//
// EmployeeDashboard renders Employee dashboard
func (app *application) EmployeeDashboard(w http.ResponseWriter, r *http.Request) {
	user := app.Session.Get(r.Context(), "user").(models.User)
	if err := app.renderTemplate(w, r, "employee-dashboard", &templateData{User: user}); err != nil {
		app.errorLog.Println(err)
	}
}

// .........Handler function for Customer Management............//

// AdminViewCustomerProfile renders list of customers or a single customer profile.
// All customer accounts are listed, if last element of path is "all". All deleted customer accounts are listed if last element of path is "deleted"
// and shows single customer profile details if last element is an int
func (app *application) AdminViewCustomerProfile(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	tmpl := "admin-view-customer-profile-list"

	data := make(map[string]interface{})
	data["profile-type"] = id

	if id != "active" && id != "deactived" && id != "deleted" && id != "all" {
		_, err := strconv.Atoi(id)
		if err != nil {
			app.errorLog.Println("Invlaid customer id")
			http.Redirect(w, r, "/page-not-found", http.StatusNotFound)
			return
		}
	}

	user := app.Session.Get(r.Context(), "user").(models.User)
	if err := app.renderTemplate(w, r, tmpl, &templateData{
		User: user,
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// .........Handler function for Transaction Management............//

// AdminViewTransaction renders list of transactions.
// All transactions are listed, if last element of path is "all".
// All refunded transactions are listed if last element of path is "refunded"
// All partially-refunded transactions are listed if last element of path is "partially-refunded"
// All pending transactions are listed if last element of path is "pending"
// All declined transactions are listed if last element of path is "declined"
// and shows single customer profile details if last element is an int
func (app *application) AdminViewTransaction(w http.ResponseWriter, r *http.Request) {
	t := path.Base(r.URL.Path)
	data := make(map[string]interface{})
	data["transaction_type"] = t

	tmpl := ""
	if t == "all" || t == "pending" || t == "cleared" || t == "declined" || t == "refunded" || t == "partially-refunded" {
		tmpl = "admin-transactions-list"
	} else if _, err := strconv.Atoi(t); err == nil {
		tmpl = "admin-transaction-details"
	} else {
		app.errorLog.Println("Invlaid customer id: ", t)
		http.Redirect(w, r, "/page-not-found", http.StatusNotFound)
		return
	}

	user := app.Session.Get(r.Context(), "user").(models.User)
	err := app.renderTemplate(w, r, tmpl, &templateData{
		User: user,
		Data: data,
	})
	if err != nil {
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

// callInvoiceMicro calls the invoice microservice to generate invoice and send it to the customer gmail
func (app *application) callInvoiceMicro(inv models.Invoice) error {
	url := "http://localhost:5000/invoice/generate-send"
	out, err := json.MarshalIndent(inv, "", "\t")
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(out))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
