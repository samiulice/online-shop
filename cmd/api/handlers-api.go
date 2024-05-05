package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"online_store/internal/cards"
	"online_store/internal/encryption"
	"online_store/internal/models"
	"online_store/internal/urlsigner"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
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

// GetDatesByID tesiting bd pool
func (app *application) GetDatesByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	dateID, _ := strconv.Atoi(id)

	d, err := app.DB.GetDate(dateID)
	if err != nil {
		app.errorLog.Println(err)
	}
	out, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		app.errorLog.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
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
		amount, _ := strconv.Atoi(data.Amount)
		txnID, err := app.SaveTransaction(models.Transaction{
			Amount:              amount,
			Currency:            data.Currency,
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
		productID, _ := strconv.Atoi(data.ProductID)
		_, err = app.SaveOrder(models.Order{
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
