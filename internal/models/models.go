package models

import (
	"time"
)

const (
	// AustraliaRegex matches Australian mobile numbers with or without country code (+61)
	AustraliaRegex = `^(\+?61|0)4\d{8}$`

	// BangladeshRegex matches Bangladeshi mobile numbers with or without country code (+880)
	BangladeshRegex = `^\+?(880)?1[3-9]\d{8}$`

	// CanadaRegex matches Canadian phone numbers in various formats
	CanadaRegex = `^(\+?1)?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`

	// FranceRegex matches French phone numbers with or without country code (+33)
	FranceRegex = `^(?:(?:\+|00)33|0)\s*[1-9](?:[\s.-]*\d{2}){4}$`

	// GermanyRegex matches German phone numbers with or without country code (+49)
	GermanyRegex = `^(\+?49|0)(\d{3,4})?[ -]?(\d{3,4})?[ -]?(\d{4,6})$`

	// IndiaRegex matches Indian mobile numbers with or without country code (+91)
	IndiaRegex = `^\+?(91)?\d{10}$`

	// JapanRegex matches Japanese phone numbers with or without country code (+81)
	JapanRegex = `^\+?81[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{4}$`

	// PakistanRegex matches Pakistani mobile numbers with or without country code (+92)
	PakistanRegex = `^\+?(92)?\d{10}$`

	// SriLankaRegex matches Sri Lankan mobile numbers with or without country code (+94)
	SriLankaRegex = `^\+?(94)?\d{9}$`

	// UKRegex matches UK phone numbers including landline, mobile, and toll-free numbers
	UKRegex = `^(?:(?:\+44\s?|0)(?:\d{5}\s?\d{5}|\d{4}\s?\d{4}\s?\d{4}|\d{3}\s?\d{3}\s?\d{4}|\d{2}\s?\d{4}\s?\d{4}|\d{4}\s?\d{4}|\d{4}\s?\d{3})|\d{5}\s?\d{4}\s?\d{4}|0800\s?\d{3}\s?\d{4})$`

	// USRegex matches US phone numbers in various formats
	USRegex = `^\+?1?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`
)

// Date is the type for all dates that holds the info about it
type Date struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`             //name of the dates
	Description     string    `json:"description"`      //description of the dates
	IsRecurring     int       `json:"is_recurring"`     //1 for subscription plan, 0 for non-recurring or one time purchase plan
	PlanID          string    `json:"plan_id"`          //stripe plan id
	PlanTitle       string    `json:"plan_title"`       //title of the plan
	PlanDescription string    `json:"plan_description"` //description of the plan
	PackageWeight   int       `json:"package_weight"`   //Weight of each package in kilogram
	PackagePrice    int       `json:"package_price"`    //Package price in USD
	StockLevel      int       `json:"stock_level"`      //number of packages in the stock
	ImageLink       string    `json:"iamge"`            //Image link of the dates
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}

// Order is the type for all orders
type Order struct {
	ID            int         `json:"id"`
	DatesID       int         `json:"dates_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerID    int         `json:"customer_id"`
	StatusID      int         `json:"status_id"` //Processing=1, Completed=2, Cancelled = 3
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Dates         Date        `json:"dates"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

// Status is the type for all order status
type Staus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TransactionStatus is the type for all transaction status
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is the type for all transaction
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	LastFourDigits      string    `json:"last_four_digits"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusID int       `json:"transaction_status_id"` //Pending=1, Cleared=2, Declined=3 Refunded=4, Partially Refunded=5
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	TransactionStatus   string    `json:"transaction_status"`
}

// TransactionData is the type for all transaction
type TransactionData struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	NameOnCard     string `json:"name_on_card"`
	Amount         int    `json:"amount"`
	Currency       string `json:"currency"`
	PaymentAmount  string `json:"payment_amount"`
	PaymentIntent  string `json:"payment_intent"`
	PaymentMethod  string `json:"payment_method"`
	LastFourDigits string `json:"last_four_digits"`
	BankReturnCode string `json:"bank_return_code"`
	ExpiryMonth    int    `json:"expiry_month"`
	ExpiryYear     int    `json:"expiry_year"`
}

// Employee is the type for Employee
type Employee struct {
	ID              int            `json:"id"`
	UserName        string         `json:"user_name"`
	FirstName       string         `json:"first_name"`
	LastName        string         `json:"last_name"`
	Gender        string         `json:"gender"`
	Address        string         `json:"address"`
	Email           string         `json:"email"`
	FacebookID      string         `json:"fb_id"`
	WhatsappID      string         `json:"whatsapp_id"`
	TwitterID       string         `json:"x_id"`
	LinkedinID      string         `json:"linkedin_id"`
	GithubID        string         `json:"github_id"`
	Mobile          string         `json:"mobile"`
	Password        string         `json:"password"`
	ImageLink       string         `json:"image_link"` //username_profile_id_yy-mm-dd_hh-mm-ss.jpg
	AccountStatusID int            `json:"account_status_id"` //Active = 1, Suspended = 2, Resigned = 3, Ex= (2 or 3)
	Credits         int            `json:"credits"`
	TaskCompleted   int            `json:"task_completed"`
	TaskCancelled   int            `json:"task_cancelled"`
	Rating          int            `json:"rating"` //5*TaskCompleted/(TaskCompleted+TaskCancelled)
	NID        string         `json:"nid"`
	NIDLink        string         `json:"nid_link"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	AccountStatus   EmployeeStatus `json:"account_status"`
}

type EmployeeStatus struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User is the type for users
type User struct {
	ID        int       `json:"id"`
	UserName  string    `json:"user_name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	ImageLink string    `json:"image_link"` //username_profile_id_yy-mm-dd_hh-mm-ss.jpf
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is the type for users
type Customer struct {
	ID            int       `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	ImageLink     string    `json:"image_link"`     //username_profile_id_yy-mm-dd_hh-mm-ss.jpf
	AccountStatus int       `json:"account_status"` //0 = deleted, 1 = active, 2 = deactivated
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ................JSON Response model for invoice microservice........................
// Order holds the necessary info to build invoice
type Invoice struct {
	ID        int              `json:"id"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Email     string           `json:"email"`
	CreatedAt time.Time        `json:"created_at"`
	Items     []InvoiceProduct `json:"items"`
}
type InvoiceProduct struct {
	ID       int    `json:"product_id"`
	Name     string `json:"product_name"`
	Quantity int    `json:"quantity"`
	Amount   int    `json:"amount"`
}

//........................................................................................
