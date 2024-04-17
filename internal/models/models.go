package models

import (
	"time"
)

// DBModel is the type for database connection values
// type DBModel struct {
// 	DB *sql.DB
// }

// Models is the wrapper for all models
// type Models struct {
// 	DB DBModel
// }

// NewModels returns Models type with database connection pool. It is used to make new Models struct in any part of tha application
// func NewModels(db *sql.DB) Models {
// 	return Models{
// 		DB: DBModel{
// 			DB: db,
// 		},
// 	}
// }

// Date is the type for all dates that holds the info about it
type Date struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	PackageSize   string    `json:"package_size"`   //size of package [Single, Family, Bulk]
	PackageWeight int       `json:"package_weight"` //Weight of each package in kilogram
	PackagePrice  int       `json:"package_price"`  //Package price in BDT
	StockLevel    int       `json:"stock_level"`    //number of packages in the stock
	Image         string    `json:"iamge"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

// Order is the type for all orders
type Order struct {
	ID            int       `json:"id"`
	DatesID       int       `json:"dates_id"`
	TransactionID int       `json:"transaction_id"`
	StatusID      int       `json:"status_id"`
	Quantity      int       `json:"quantity"`
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
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
	LastFourDigits      string    `json:"last_four_digits"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusID int       `json:"transaction_status_id"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// User is the type for users
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
