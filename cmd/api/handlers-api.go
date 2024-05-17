package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

	"github.com/stripe/stripe-go"
	"golang.org/x/crypto/bcrypt"
)

type stripePayload struct {
	PlanID         string `json:"plan_id"`
	ProductID      string `json:"product_id"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	PaymentIntent  string `json:"payment_intent"`
	PaymentMethod  string `json:"payment_method"`
	CardBrand      string `json:"card_brand"`
	LastFourDigits string `json:"last_four_digits"`
	ExpiryMonth    int    `json:"expiry_month"`
	ExpiryYear     int    `json:"expiry_year"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {

	var payload stripePayload
	err := json.NewDecoder(r.Body).Decode((&payload))
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true

	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		okay = false
	}

	if okay {
		out, err := json.MarshalIndent(pi, "", "    ")
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {
		j := jsonResponse{
			OK:      false,
			Message: msg,
			Content: "",
		}

		out, err := json.MarshalIndent(j, "", "    ")
		if err != nil {
			app.errorLog.Println(err)
		}

		w.Header().Set("Content_Type", "application/json")
		w.Write(out)
	}
}

func (app *application) CreateCustomerAndSubscribe(w http.ResponseWriter, r *http.Request) {
	var data stripePayload
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: data.Currency,
	}

	ok := true
	var subscription *stripe.Subscription
	txnMsg := "Transaction successfull"

	stripeCustomer, msg, err := card.CreateCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.errorLog.Println(err)
		ok = false
		txnMsg = msg
	}

	if ok {
		subscription, err = card.Subscribe(stripeCustomer, data.PlanID, data.Email, data.LastFourDigits, "")
		if err != nil {
			app.errorLog.Println(err)
			ok = false
			txnMsg = "Error subscribing customer"
		}
		app.infoLog.Println("Subscription id : ", subscription.ID)
	}

	var orderID, productID int
	amount, _ := strconv.Atoi(data.Amount)
	productID, _ = strconv.Atoi(data.ProductID)
	if ok {
		//save customer info for each transaction
		customerID, err := app.SaveCustomer(models.Customer{
			FirstName: data.FirstName,
			LastName:  data.LastName,
			Email:     data.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			app.infoLog.Println(err)
			return
		}
		//save transaction info to the database
		txnID, err := app.SaveTransaction(models.Transaction{
			Amount:              amount,
			Currency:            data.Currency,
			PaymentIntent:       subscription.ID,
			PaymentMethod:       data.PaymentMethod,
			LastFourDigits:      data.LastFourDigits,
			TransactionStatusID: 2,
			ExpiryMonth:         data.ExpiryMonth,
			ExpiryYear:          data.ExpiryYear,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		})
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		//save order to database
		orderID, err = app.SaveOrder(models.Order{
			DatesID:       productID,
			CustomerID:    customerID,
			TransactionID: txnID,
			StatusID:      1,
			Quantity:      1,
			Amount:        amount,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}

	//call invoice microservice to generate invoice template and send it to the customer email address
	var product = models.InvoiceProduct{
		ID:       orderID,
		Name:     "Ajwa",
		Quantity: 1,
		Amount:   amount,
	}
	var items = []models.InvoiceProduct{product}
	var inv = models.Invoice{
		ID:        productID,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		CreatedAt: time.Now(),
		Items:     items,
	}
	err = app.callInvoiceMicro(inv)
	if err != nil {
		app.errorLog.Println(err)
	}

	resp := jsonResponse{
		OK:      ok,
		Message: txnMsg,
	}

	out, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.badRequest(w, err)
	}

	//get the user from the database by username; send error if invalid username
	user, err := app.DB.GetUserDetails(userInput.UserName, "user_name")
	if err != nil {
		app.invalidCradentials(w)
		return
	}

	//validate the password; send error if invalid password
	validPassword, err := app.passwordMatchers(user.Password, userInput.Password)
	if err != nil || !validPassword {
		app.invalidCradentials(w)
		return
	}

	//generate the token
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.badRequest(w, err)
	}

	//save the token to database
	err = app.DB.InsertToken(token, user)
	if err != nil {
		app.badRequest(w, err)
	}

	//Add id to the session

	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
		UserID  int           `json:"user_id"`
	}
	payload.Error = false
	payload.Message = "token generated"
	payload.Token = token
	payload.UserID = user.ID

	//send response
	err = app.writeJSON(w, http.StatusOK, payload)
	if err != nil {
		app.infoLog.Println(err)
	}
}

func (app *application) CheckAuthenticated(w http.ResponseWriter, r *http.Request) {
	//validate the token and get associated user
	user, err := app.authenticateToken(r)
	if err != nil {
		app.invalidCradentials(w)
		return
	}

	//valid user
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = false
	payload.Message = "authenticated user " + user.UserName
	app.writeJSON(w, http.StatusOK, payload)

}

// ForgotPassword facilitates reset password mechanism for registered user
func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		UserName string `json:"user_name"`
	}
	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	//verify the user
	user, err := app.DB.GetUserDetails(userInput.UserName, "user_name")
	if err != nil {
		response.Error = true
		response.Message = "Username or Email doesn't match"
		app.writeJSON(w, http.StatusAccepted, response)
		return
	}

	//Generate signed url
	link := fmt.Sprintf("%s/reset-password?email=%s&user_id=%v", app.config.frontend, user.Email, user.ID)
	sign := urlsigner.Signer{
		Secret: []byte(app.config.secretKey),
	}
	signedLink := sign.GenerateTokenFromString(link)

	//Send Mail
	var data struct {
		Link string `json:"link"`
	}

	data.Link = signedLink

	err = app.SendMail("info@demomailtrap.com", "coding.samiul@gmail.com", "Password reset request", "reset-password", data)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, err)
		return
	}

	//Send JSON Response to the frontend after sending email successfully
	response.Error = false
	response.Message = "A password reset link sent to your email"
	app.writeJSON(w, http.StatusOK, response)

}

// ResetPassword saves the newly entered password to the database
func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		ID                 string `json:"user_id"`
		Email              string `json:"email"`
		NewPassword        string `json:"new_password"`
		ConfirmNewPassword string `json:"confirm_new_password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	//Decrypt email and ID
	decryptor := encryption.Encryption{
		Key: []byte(app.config.secretKey),
	}

	_, err = decryptor.Decrypt(userInput.Email)
	if err != nil {
		app.errorLog.Println("falied to decrypt email:\t", err)
		return
	}
	decryptedUserID, err := decryptor.Decrypt(userInput.ID)
	if err != nil {
		app.errorLog.Println("falied to decrypt userID:\t", err)
		return
	}

	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	//Varifay that two password are same
	if userInput.NewPassword != userInput.ConfirmNewPassword {
		response.Error = true
		response.Message = "Password mismatch"
		app.writeJSON(w, http.StatusAccepted, response)
		return
	}

	//update password
	newhash, err := bcrypt.GenerateFromPassword([]byte(userInput.NewPassword), 12)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	err = app.DB.UpdatePasswordByUserID(decryptedUserID, string(newhash))
	if err != nil {
		app.badRequest(w, err)
		return
	}

	//Send JSON Response to the frontend after sending email successfully
	response.Error = false
	response.Message = "Password updated. Redirecting..."
	app.writeJSON(w, http.StatusOK, response)

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

func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	var txnData models.TransactionData
	err := app.readJSON(w, r, &txnData)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.secret,
	}

	pi, err := card.RetrivePaymentIntent(txnData.PaymentIntent)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	pm, err := card.GetPaymentMethod(txnData.PaymentMethod)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	//Fill txnData
	txnData.LastFourDigits = pm.Card.Last4
	txnData.BankReturnCode = pi.Charges.Data[0].ID
	txnData.ExpiryMonth = int(pm.Card.ExpMonth)
	txnData.ExpiryYear = int(pm.Card.ExpYear)
	//amount, currency, payment_intent, payment_method, last_four_digits,
	//  bank_return_code, transaction_status_id, expiry_month, expiry_year, created_at, updated_at)

	txn := models.Transaction{
		Amount:              txnData.Amount,
		Currency:            txnData.Currency,
		PaymentIntent:       txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		LastFourDigits:      txnData.LastFourDigits,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
	}
	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	app.writeJSON(w, http.StatusOK, txn)
}


// GetOrdersHistoy return list of all sales to the corresponded category in JSON format
func (app *application) GetOrdersHistoy(w http.ResponseWriter, r *http.Request) {
	statusType := path.Base(r.URL.Path)
	if statusType[0] >= '0' && statusType[0] <= '9' {
		Orders, err := app.DB.GetOrdersHistory(statusType)
		if err != nil {
			app.errorLog.Println(err)
			app.badRequest(w, err)
			return
		}
		app.writeJSON(w, http.StatusOK, Orders)
	} else {
		var payload struct {
			PageSize         int `json:"page_size"`
			CurrentPageIndex int `json:"current_page_index"`
		}
		err := app.readJSON(w, r, &payload)
		if err != nil {
			app.badRequest(w, err)
			return
		}
		Orders, totalRecords, err := app.DB.GetOrdersHistoryPaginated(statusType, payload.PageSize, payload.CurrentPageIndex)

		if err != nil {
			app.errorLog.Println(err)
			app.badRequest(w, err)
			return
		}
		var Resp struct {
			PageSize         int             `json:"page_size"`
			CurrentPageIndex int             `json:"current_page_index"`
			TotalRecords     int             `json:"total_records"`
			Orders           []*models.Order `json:"orders"`
		}
		Resp.PageSize = payload.PageSize
		Resp.CurrentPageIndex = payload.CurrentPageIndex
		Resp.TotalRecords = totalRecords
		Resp.Orders = Orders
		app.writeJSON(w, http.StatusOK, Resp)
	}

}

// GetTransactionHistory return list of all sales to the corresponded category in JSON format
func (app *application) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	statusType := path.Base(r.URL.Path)

	if statusType[0] >= '0' && statusType[0] <= '9' {
		Transactions, err := app.DB.GetTransactionsHistory(statusType)
		if err != nil {
			app.errorLog.Println(err)
			app.badRequest(w, err)
			return
		}
		app.writeJSON(w, http.StatusOK, Transactions)
	} else {
		var payload struct {
			PageSize         int `json:"page_size"`
			CurrentPageIndex int `json:"current_page_index"`
		}
		err := app.readJSON(w, r, &payload)
		if err != nil {
			app.badRequest(w, err)
			return
		}
		Transactions, totalRecords, err := app.DB.GetTransactionsHistoryPaginated(statusType, payload.PageSize, payload.CurrentPageIndex)

		if err != nil {
			app.errorLog.Println(err)
			app.badRequest(w, err)
			return
		}
		var Resp struct {
			PageSize         int                   `json:"page_size"`
			CurrentPageIndex int                   `json:"current_page_index"`
			TotalRecords     int                   `json:"total_records"`
			Transactions     []*models.Transaction `json:"transactions"`
		}
		Resp.PageSize = payload.PageSize
		Resp.CurrentPageIndex = payload.CurrentPageIndex
		Resp.TotalRecords = totalRecords
		Resp.Transactions = Transactions
		app.writeJSON(w, http.StatusOK, Resp)
	}
}
// ManageEmployeeAccount manages employee account
func (app *application) ManageEmployeeAccount(w http.ResponseWriter, r *http.Request) {
	

	url := strings.Split(r.URL.Path, "/")
	action := url[5]
	id, err := strconv.Atoi(url[6])
	if err != nil {
		app.badRequest(w, err)
		return
	}
	var msg string
	if action == "activate" {
		err = app.DB.UpdateEmployeeAccountStatusByID(id, 1)
		msg = "Account renewed and activated..."
	} else if action == "suspend" {
		err = app.DB.UpdateEmployeeAccountStatusByID(id, 2)
		msg = "Account suspened..."
	} else if action == "revoke" {
		err = app.DB.UpdateEmployeeAccountStatusByID(id, 3)
		msg = "Account revoked and disabled..."
	} else if action == "rejoin"{
		err = app.DB.UpdateEmployeeAccountStatusByID(id, 1)
		msg = "Account rejoined and enabled..."
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	
	if err != nil {
		app.errorLog.Println(err)
		resp.Error = true
		resp.Message = "Unable to perform this action! please try again"
		app.writeJSON(w, http.StatusOK, resp)
		return
	}
	
	resp.Error = false
	resp.Message = msg
	app.writeJSON(w, http.StatusOK, resp)

}
// GetEmployeeList return list of employees to the corresponded category in JSON format
func (app *application) GetEmployees(w http.ResponseWriter, r *http.Request) {
	accountType := path.Base(r.URL.Path)

	id, err := strconv.Atoi(accountType)
	if err == nil {
		employee, err := app.DB.GetEmployeeByID(id)
		if err != nil {
			app.errorLog.Println(err)
			app.badRequest(w, err)
			return
		}
		app.writeJSON(w, http.StatusOK, employee)
	} else {
		var payload struct {
			PageSize         int `json:"page_size"`
			CurrentPageIndex int `json:"current_page_index"`
		}
		err := app.readJSON(w, r, &payload)
		if err != nil {
			app.badRequest(w, err)
			return
		}
		employees, totalRecords, err := app.DB.GetEmployeeListPaginated(accountType, payload.PageSize, payload.CurrentPageIndex)

		if err != nil {
			app.errorLog.Println(err)
			app.badRequest(w, err)
			return
		}
		var Resp struct {
			PageSize         int                   `json:"page_size"`
			CurrentPageIndex int                   `json:"current_page_index"`
			TotalRecords     int                   `json:"total_records"`
			Employees     []*models.Employee `json:"employees"`
		}
		Resp.PageSize = payload.PageSize
		Resp.CurrentPageIndex = payload.CurrentPageIndex
		Resp.TotalRecords = totalRecords
		Resp.Employees = employees
		app.writeJSON(w, http.StatusOK, Resp)
	}
}

// AdminCustomerProfile return list of all customer in JSON format
func (app *application) AdminCustomerProfile(w http.ResponseWriter, r *http.Request) {

	id := strings.Split(r.RequestURI, "/")[6]
	Sales, err := app.DB.GetCustomerProfile(id)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, err)
		return
	}
	app.writeJSON(w, http.StatusOK, Sales)
}

// RefundCharge refund the charged money to the customer account
func (app *application) RefundCharge(w http.ResponseWriter, r *http.Request) {
	lastPart := path.Base(r.URL.Path)
	ids := strings.Split(lastPart, "-")

	orderStatusID, err := strconv.Atoi(ids[0])
	if err != nil {
		app.badRequest(w, err)
		return
	}
	transactionStatusID, err := strconv.Atoi(ids[1])
	if err != nil {
		app.badRequest(w, err)
		return
	}

	tr, err := app.DB.GetTransactionsHistory(ids[1])
	if err != nil {
		app.badRequest(w, err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: tr[0].Currency,
	}

	err = card.Refund(tr[0].PaymentIntent, tr[0].Amount)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	//update database
	err = app.DB.UpdateTransactionStatusID(transactionStatusID, 4) //update status id = 4 for refunded order
	if err != nil {
		app.badRequest(w, errors.New("order refunded suceessfully, but unable to update transaction status"))
		return
	}
	err = app.DB.UpdateOrderStatusID(orderStatusID, 3) //update status id = 3 for cancelled order
	if err != nil {
		app.badRequest(w, errors.New("order refunded suceessfully, but unable to update transaction status"))
		return
	}
	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "Order Refunded Successfully"
	app.writeJSON(w, http.StatusOK, resp)
}

// CancelSubscription cancel a subscription
func (app *application) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	lastPart := path.Base(r.URL.Path)
	ids := strings.Split(lastPart, "-")

	orderStatusID, err := strconv.Atoi(ids[0])
	if err != nil {
		app.badRequest(w, err)
		return
	}
	transactionStatusID, err := strconv.Atoi(ids[1])
	if err != nil {
		app.badRequest(w, err)
		return
	}

	tr, err := app.DB.GetTransactionsHistory(ids[1])
	if err != nil {
		app.badRequest(w, err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: tr[0].Currency,
	}

	err = card.CancelSubscription(tr[0].PaymentIntent)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	//update database
	err = app.DB.UpdateTransactionStatusID(transactionStatusID, 4) //update status id = 4 for refunded order
	if err != nil {
		app.badRequest(w, errors.New("subscription Cancelled suceessfully, but unable to update transaction status"))
		return
	}
	err = app.DB.UpdateOrderStatusID(orderStatusID, 3) //update status id = 3 for cancelled order
	if err != nil {
		app.badRequest(w, errors.New("subscription Cancelled suceessfully, but unable to update order status"))

		return
	}
	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "Subscription Cancelled Successfully"
	app.writeJSON(w, http.StatusOK, resp)
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
