package dbrepo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"online_store/internal/models"
	"strconv"
	"time"
)

//IsRegistered chceks whether an user data is already exist or not
//If exist then return true, id, nil
//If doesn't exist then return false, 0, nil
//If error occured then return false, 0, error
func(p *postgresDBRepo) IsRegistered(userType, paramType, paramValue string) (int, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	query := fmt.Sprintf("SELECT id	FROM %s	WHERE %s = $1", userType, paramType) 
	err := p.DB.QueryRowContext(ctx, query, paramValue).Scan(&id)

	if err == sql.ErrNoRows {
		// does not exist
		return 0, nil
	} else if err != nil {
		//Database error
		return 0, err
	}

	return id, nil
	
}

//VerifyUser chceks user validity
//If valid user then return id, email, mobile, nil
//If doesn't exist then return 0,"", "", nil
//If error occured then return 0,"", "", error
func(p *postgresDBRepo) VerifyUser(userType, searchParam, paramValue string) (int, string, string, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var email, mobile string
	query := fmt.Sprintf("SELECT id, email, mobile	FROM %s	WHERE %s = $1", userType, searchParam) 
	err := p.DB.QueryRowContext(ctx, query, paramValue).Scan(&id, &email, &mobile)

	return id, email, mobile, err
	
}

// GetDate return a date package for specific id
func (p *postgresDBRepo) GetDate(id int) (models.Date, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var d models.Date

	query := `
		SELECT id, name, description, is_recurring, plan_id, plan_title, plan_description, 
			package_weight, package_price, stock_level, coalesce(image_link, ''), created_at, updated_at
		FROM dates
		WHERE id = $1`
	row := p.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&d.ID,
		&d.Name,
		&d.Description,
		&d.IsRecurring,
		&d.PlanID,
		&d.PlanTitle,
		&d.PlanDescription,
		&d.PackageWeight,
		&d.PackagePrice,
		&d.StockLevel,
		&d.ImageLink,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	return d, err
}

// InsertTransaction inserts new transaction to the database and returns its id
func (p *postgresDBRepo) InsertTransaction(txn models.Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO transactions (amount, currency, payment_intent, payment_method, last_four_digits, bank_return_code, transaction_status_id, expiry_month, expiry_year, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id
	`

	var id int
	err := p.DB.QueryRowContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.PaymentIntent,
		txn.PaymentMethod,
		txn.LastFourDigits,
		txn.BankReturnCode,
		txn.TransactionStatusID,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// InsertOrder inserts new order to the database and returns its id
func (p *postgresDBRepo) InsertOrder(order models.Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO orders (dates_id, transaction_id, customer_id, status_id, quantity, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id
	`

	var id int
	err := p.DB.QueryRowContext(ctx, stmt,
		order.DatesID,
		order.TransactionID,
		order.CustomerID,
		order.StatusID,
		order.Quantity,
		order.Amount,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// InsertCustomer inserts new customer to the database and returns its id
func (p *postgresDBRepo) InsertCustomer(customer models.Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO customers (user_name, first_name, last_name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) returning id
	`

	var id int
	err := p.DB.QueryRowContext(ctx, stmt,
		customer.UserName,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// Database Functions that is related to Order processing activity
// GetOrdersHistory returns a slice of all orders with associated customer and transaction info.
//if statusType == all, it will return list all orders
//if statusType == completed, it will return list of completed orders
//if statusType == refunded, it will return list of refunded orders
//if statusType == cancelled, it will return list of cancelled orders

func (p *postgresDBRepo) GetOrdersHistory(statusType string) ([]*models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var orders []*models.Order

	query := `
		SELECT 
			o.id, o.dates_id, o.transaction_id, o.customer_id, o.status_id, o.quantity,
			o.amount, o.created_at, o.updated_at, d.id, d.name, d.description, d.is_recurring,
			d.plan_id, d.plan_title, d.plan_description, d.package_weight, d.package_price,
			d.stock_level, d.image_link, d.created_at, d.updated_at, t.id, t.amount, t.currency,
			t.payment_intent, t.payment_method, t.last_four_digits, t.bank_return_code, 
			t.transaction_status_id, t.expiry_month, t.expiry_year, t.created_at, t.updated_at,
			 ts.name, c.id, c.user_name, c.first_name, c.last_name, c.email, c.image_link, c.created_at, c.updated_at
		FROM
			orders o
			LEFT JOIN dates d on (o.dates_id = d.id)
			LEFT JOIN transactions t on (o.transaction_id = t.id)
			LEFT JOIN transaction_status ts on (t.transaction_status_id = ts.id)
			LEFT JOIN customers c on (o.customer_id = c.id)
		`

	trails := `
		 ORDER BY
			o.created_at desc`

	var rows *sql.Rows
	var err error

	if statusType == "all" {
		query = query + trails
	} else if statusType == "processing" {
		query = query + ` WHERE o.status_id = 1` + trails
	} else if statusType == "completed" {
		query = query + ` WHERE o.status_id = 2` + trails
	} else if statusType == "cancelled" {
		query = query + ` WHERE o.status_id = 3` + trails
	} else if statusType == "pending" {
		query = query + ` WHERE t.transaction_status_id = 1` + trails
	} else if statusType == "cleared" {
		query = query + ` WHERE t.transaction_status_id = 2` + trails
	} else if statusType == "declined" {
		query = query + ` WHERE t.transaction_status_id = 3` + trails
	} else if statusType == "refunded" {
		query = query + ` WHERE t.transaction_status_id = 4` + trails
	} else if statusType == "partially-refunded" {
		query = query + ` WHERE t.transaction_status_id = 5` + trails
	} else if statusType == "one-off" {
		query = query + ` WHERE d.is_recurring = 0` + trails
	} else if statusType == "subscriptions" {
		query = query + ` WHERE d.is_recurring = 1` + trails
	} else if _, err := strconv.Atoi(statusType); err == nil {
		query = query + ` WHERE o.id = ` + statusType + trails
	} else {
		return orders, errors.New("invalid function parameter for the database function call")
	}

	rows, err = p.DB.QueryContext(ctx, query)
	if err != nil {
		return orders, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Order
		err = rows.Scan(
			&o.ID,
			&o.DatesID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,

			&o.Dates.ID,
			&o.Dates.Name,
			&o.Dates.Description,
			&o.Dates.IsRecurring,
			&o.Dates.PlanID,
			&o.Dates.PlanTitle,
			&o.Dates.PlanDescription,
			&o.Dates.PackageWeight,
			&o.Dates.PackagePrice,
			&o.Dates.StockLevel,
			&o.Dates.ImageLink,
			&o.Dates.CreatedAt,
			&o.Dates.UpdatedAt,

			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.PaymentIntent,
			&o.Transaction.PaymentMethod,
			&o.Transaction.LastFourDigits,
			&o.Transaction.BankReturnCode,
			&o.Transaction.TransactionStatusID,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.CreatedAt,
			&o.Transaction.UpdatedAt,
			&o.Transaction.TransactionStatus,

			&o.Customer.ID,
			&o.Customer.UserName,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
			&o.Customer.ImageLink,
			&o.Customer.CreatedAt,
			&o.Customer.UpdatedAt,
		)
		if err != nil {
			return orders, err
		}
		orders = append(orders, &o)
	}

	// return orders, errors.New("testing errors")
	return orders, nil
}

func (p *postgresDBRepo) GetOrdersHistoryPaginated(statusType string, pageSize, currentPageIndex int) ([]*models.Order, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (currentPageIndex - 1) * pageSize

	var orders []*models.Order

	query := `
		SELECT 
			o.id, o.dates_id, o.transaction_id, o.customer_id, o.status_id, o.quantity,
			o.amount, o.created_at, o.updated_at, d.id, d.name, d.description, d.is_recurring,
			d.plan_id, d.plan_title, d.plan_description, d.package_weight, d.package_price,
			d.stock_level, d.image_link, d.created_at, d.updated_at, t.id, t.amount, t.currency,
			t.payment_intent, t.payment_method, t.last_four_digits, t.bank_return_code, 
			t.transaction_status_id, t.expiry_month, t.expiry_year, t.created_at, t.updated_at,
			 ts.name, c.id, c.user_name, c.first_name, c.last_name, c.email, c.image_link, c.created_at, c.updated_at
		FROM
			orders o
			LEFT JOIN dates d on (o.dates_id = d.id)
			LEFT JOIN transactions t on (o.transaction_id = t.id)
			LEFT JOIN transaction_status ts on (t.transaction_status_id = ts.id)
			LEFT JOIN customers c on (o.customer_id = c.id)
		`

	trails := `
		 ORDER BY
			t.id asc
		LIMIT $1 OFFSET $2
		`

	newQuery := `
		SELECT 
			COUNT(o.id)
		FROM
			orders o
			LEFT JOIN dates d on (o.dates_id = d.id)
			LEFT JOIN transactions t on (o.transaction_id = t.id)
			LEFT JOIN transaction_status ts on (t.transaction_status_id = ts.id)
			LEFT JOIN customers c on (o.customer_id = c.id)
		`
	var rows *sql.Rows
	var err error

	if statusType == "all" {
		query = query + trails
	} else if statusType == "processing" {
		query = query + ` WHERE o.status_id = 1` + trails
		newQuery = newQuery + ` WHERE o.status_id = 1`
	} else if statusType == "completed" {
		query = query + ` WHERE o.status_id = 2` + trails
		newQuery = newQuery + ` WHERE o.status_id = 2`
	} else if statusType == "cancelled" {
		query = query + ` WHERE o.status_id = 3` + trails
		newQuery = newQuery + ` WHERE o.status_id = 3`
	} else if statusType == "pending" {
		query = query + ` WHERE t.transaction_status_id = 1` + trails
		newQuery = newQuery + ` WHERE t.transaction_status_id = 1`
	} else if statusType == "cleared" {
		query = query + ` WHERE t.transaction_status_id = 2` + trails
		newQuery = newQuery + ` WHERE t.transaction_status_id = 2`
	} else if statusType == "declined" {
		query = query + ` WHERE t.transaction_status_id = 3` + trails
		newQuery = newQuery + ` WHERE t.transaction_status_id = 3`
	} else if statusType == "refunded" {
		query = query + ` WHERE t.transaction_status_id = 4` + trails
		newQuery = newQuery + ` WHERE t.transaction_status_id = 4`
	} else if statusType == "partially-refunded" {
		query = query + ` WHERE t.transaction_status_id = 5` + trails
		newQuery = newQuery + ` WHERE t.transaction_status_id = 5`
	} else if statusType == "one-off" {
		query = query + ` WHERE d.is_recurring = 0` + trails
		newQuery = newQuery + ` WHERE d.is_recurring = 0`
	} else if statusType == "subscriptions" {
		query = query + ` WHERE d.is_recurring = 1` + trails
		newQuery = newQuery + ` WHERE d.is_recurring = 1`
	}

	rows, err = p.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return orders, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Order
		err = rows.Scan(
			&o.ID,
			&o.DatesID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,

			&o.Dates.ID,
			&o.Dates.Name,
			&o.Dates.Description,
			&o.Dates.IsRecurring,
			&o.Dates.PlanID,
			&o.Dates.PlanTitle,
			&o.Dates.PlanDescription,
			&o.Dates.PackageWeight,
			&o.Dates.PackagePrice,
			&o.Dates.StockLevel,
			&o.Dates.ImageLink,
			&o.Dates.CreatedAt,
			&o.Dates.UpdatedAt,

			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.PaymentIntent,
			&o.Transaction.PaymentMethod,
			&o.Transaction.LastFourDigits,
			&o.Transaction.BankReturnCode,
			&o.Transaction.TransactionStatusID,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.CreatedAt,
			&o.Transaction.UpdatedAt,
			&o.Transaction.TransactionStatus,

			&o.Customer.ID,
			&o.Customer.UserName,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
			&o.Customer.ImageLink,
			&o.Customer.CreatedAt,
			&o.Customer.UpdatedAt,
		)
		if err != nil {
			return orders, 0, err
		}
		orders = append(orders, &o)
	}

	var totalRecords int
	countRow := p.DB.QueryRowContext(ctx, newQuery)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return orders, 0, err
	}
	return orders, totalRecords, nil
}

func (p *postgresDBRepo) UpdateOrderStatusID(id, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE orders
		SET status_id = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := p.DB.ExecContext(ctx, stmt, statusID, time.Now(), id)

	return err
}

// Database Functions that is related to Order processing activity
// GetOrdersHistory returns a slice of all orders with associated customer and transaction info.
//if statusType == all, it will return list all orders
//if statusType == completed, it will return list of completed orders
//if statusType == refunded, it will return list of refunded orders
//if statusType == cancelled, it will return list of cancelled orders

func (p *postgresDBRepo) GetTransactionsHistory(statusType string) ([]*models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var transactions []*models.Transaction

	query := `
		SELECT 
			t.id, t.amount, t.currency, t.payment_intent, t.payment_method,
			t.last_four_digits, t.bank_return_code, t.transaction_status_id, t.expiry_month, t.expiry_year,
			t.created_at, t.updated_at, ts.name
		FROM
			transactions t
			LEFT JOIN transaction_status ts on (t.transaction_status_id = ts.id)
		`

	trails := `
		 ORDER BY
			t.created_at desc`

	var rows *sql.Rows
	var err error

	if statusType == "all" {
		query = query + trails
	} else if statusType == "pending" {
		query = query + ` WHERE t.transaction_status_id = 1` + trails
	} else if statusType == "cleared" {
		query = query + ` WHERE t.transaction_status_id = 2` + trails
	} else if statusType == "declined" {
		query = query + ` WHERE t.transaction_status_id = 3` + trails
	} else if statusType == "refunded" {
		query = query + ` WHERE t.transaction_status_id = 4` + trails
	} else if statusType == "partially-refunded" {
		query = query + ` WHERE t.transaction_status_id = 5` + trails
	} else {
		query = query + ` WHERE t.id = ` + statusType + trails
	}

	rows, err = p.DB.QueryContext(ctx, query)
	if err != nil {
		return transactions, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Transaction
		err = rows.Scan(
			&t.ID,
			&t.Amount,
			&t.Currency,
			&t.PaymentIntent,
			&t.PaymentMethod,
			&t.LastFourDigits,
			&t.BankReturnCode,
			&t.TransactionStatusID,
			&t.ExpiryMonth,
			&t.ExpiryYear,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.TransactionStatus,
		)
		if err != nil {
			return transactions, err
		}
		transactions = append(transactions, &t)
	}

	// return orders, errors.New("testing errors")
	return transactions, nil
}
func (p *postgresDBRepo) GetTransactionsHistoryPaginated(statusType string, pageSize, currentPageIndex int) ([]*models.Transaction, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	offset := (currentPageIndex - 1) * pageSize
	var transactions []*models.Transaction

	query := `
		SELECT 
			t.id, t.amount, t.currency, t.payment_intent, t.payment_method,
			t.last_four_digits, t.bank_return_code, t.transaction_status_id, t.expiry_month, t.expiry_year,
			t.created_at, t.updated_at, ts.name
		FROM
			transactions t
			LEFT JOIN transaction_status ts on (t.transaction_status_id = ts.id)
		`

	trails := `
		 ORDER BY
			t.id asc
		LIMIT $1 OFFSET $2`
	newQuery := `
	SELECT 
		COUNT(t.id)
	FROM
		transactions t
		LEFT JOIN transaction_status ts on (t.transaction_status_id = ts.id)
	`
	var rows *sql.Rows
	var err error

	if statusType == "all" {
		query = query + trails
	} else if statusType == "pending" {
		query += ` WHERE t.transaction_status_id = 1` + trails
		newQuery += ` WHERE t.transaction_status_id = 1`
	} else if statusType == "cleared" {
		query += ` WHERE t.transaction_status_id = 2` + trails
		newQuery += ` WHERE t.transaction_status_id = 2`
	} else if statusType == "declined" {
		query += ` WHERE t.transaction_status_id = 3` + trails
		newQuery += ` WHERE t.transaction_status_id = 3`
	} else if statusType == "refunded" {
		query += ` WHERE t.transaction_status_id = 4` + trails
		newQuery += ` WHERE t.transaction_status_id = 4`
	} else if statusType == "partially-refunded" {
		query += ` WHERE t.transaction_status_id = 5` + trails
		newQuery += ` WHERE t.transaction_status_id = 5`
	}

	rows, err = p.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return transactions, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Transaction
		err = rows.Scan(
			&t.ID,
			&t.Amount,
			&t.Currency,
			&t.PaymentIntent,
			&t.PaymentMethod,
			&t.LastFourDigits,
			&t.BankReturnCode,
			&t.TransactionStatusID,
			&t.ExpiryMonth,
			&t.ExpiryYear,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.TransactionStatus,
		)
		if err != nil {
			return transactions, 0, err
		}
		transactions = append(transactions, &t)
	}

	var totalRecords int
	countRow := p.DB.QueryRowContext(ctx, newQuery)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return transactions, 0, err
	}
	return transactions, totalRecords, nil
}

// UpdateTransactionStatus update the status id for a transaction
func (p *postgresDBRepo) UpdateTransactionStatusID(id, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE transactions
		SET transaction_status_id = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := p.DB.ExecContext(ctx, stmt, statusID, time.Now(), id)

	return err
}

// Database Functions that relates to User Account activity
// GetUserbyUserName gets a user by userName
func (p *postgresDBRepo) GetUserDetails(index, paramType, account_type string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//paramType : ID >>> id
	//paramType : username >>> user_name
	//paramType : (contains @) >>> email
	var u models.User

	query := `
		SELECT id, user_name, first_name, last_name, email, password, coalesce(image_link, ''), created_at, updated_at
		FROM ` + account_type + `
		 WHERE ` + paramType + ` = $1`
	row := p.DB.QueryRowContext(ctx, query, index)

	err := row.Scan(
		&u.ID,
		&u.UserName,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.ImageLink,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return u, err
}

// UpdateUserPasswordByID updates account password for a user
func (p *postgresDBRepo) UpdateUserPasswordByID(userType, id, newPassword string) error{
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := fmt.Sprintf("UPDATE	%s SET password=$1	WHERE id=$2", userType)

	_, err := p.DB.ExecContext(ctx, stmt, newPassword, id)

	return err
}

// Function that relates to the Token
//GetUserInitialData returns models.UserInitialData for generating token
func (p *postgresDBRepo) GetUserInitialData(userName, param, tableName string)(models.UserInitialData, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var u models.UserInitialData
	query := `
	SELECT id, first_name, email, password, image_link, account_status_id, created_at, updated_at
	FROM ` + tableName + ` WHERE ` + param + `= $1`

	row := p.DB.QueryRowContext(ctx, query, userName)

	err := row.Scan(
		&u.UserID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.ImageLink,
		&u.AccountStatusID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	return u, err
}


// InsertToken inserts token to database
func (p *postgresDBRepo) InsertToken(t *models.Token, u models.UserInitialData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Delete existing tokens for the user
	stmt := `DELETE FROM tokens WHERE user_id = $1`
	_, err := p.DB.ExecContext(ctx, stmt, u.UserID)
	if err != nil {
		return err
	}

	stmt = `
		INSERT INTO tokens (user_id, name, email, acc_type, token_hash, expiry, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = p.DB.ExecContext(ctx, stmt, //not bothering about the result
		u.UserID,
		u.Name,
		u.Email,
		u.AccountType,
		t.Hash,
		t.Expiry,
		time.Now(),
		time.Now(),
	)

	return err
}

// GetUserbyToken returns user info from tokens table
func (p *postgresDBRepo) GetUserbyToken(token string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var u models.User

	accountType, err := p.GetAccountTypeByToken(token)
	if err != nil {
		return nil, err
	}

	query := `
			SELECT u.id, u.first_name, u.last_name, u.email
			FROM ` + accountType + ` u
				INNER JOIN tokens t ON (u.id = t.user_id)
			WHERE t.token_hash = $1
				AND t.expiry > $2
	`

	row := p.DB.QueryRowContext(ctx, query, tokenHash[:], time.Now())

	err = row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAccountTypeByToken returns account type of particular user from tokens table
func (p *postgresDBRepo) GetAccountTypeByToken(token string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))

	var accountType string
	var err error

	query := `
			SELECT acc_type 
			FROM tokens 
			WHERE token_hash = $1`

	row := p.DB.QueryRowContext(ctx, query, tokenHash[:])

	err = row.Scan(
		&accountType,
	)
	if err != nil {
		return accountType, err
	}
	return accountType, nil
}

// Database Functions that is related to Customer profile

// GetCustomerProfile returns a slice of all customer's info.
// if index == all, it will return list all profiles
// if index == deleted, it will return list of deleted profiles
// if index == active, it will return list of active profiles
// if index == deactive, it will return list of deactive profiles
// if index is a type of int, it will return customer profile corresponds to id = index
func (p *postgresDBRepo) GetCustomerProfile(index string) ([]*models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var customers []*models.Customer

	query := `
		SELECT 
			id, user_name, first_name, last_name, email, image_link, account_status, created_at, updated_at
		FROM
			customers
		`

	trails := `
		 ORDER BY
			created_at desc`

	if index == "all" {
		query = query + trails
	} else if index == "deleted" {
		query = query + ` WHERE account_status = 0` + trails
	} else if index == "active" {
		query = query + ` WHERE account_status = 1` + trails
	} else if index == "deactived" {
		query = query + ` WHERE account_status = 2` + trails
	} else {
		query = query + ` WHERE id = ` + index + trails

	}

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return customers, err
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Customer
		err = rows.Scan(
			&c.ID,
			&c.UserName,
			&c.FirstName,
			&c.LastName,
			&c.Email,
			&c.ImageLink,
			&c.AccountStatusID,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return customers, err
		}
		customers = append(customers, &c)
	}

	// return orders, errors.New("testing errors")
	return customers, nil
}

// Database Functions that is related to Employee Account Activity

//UserPreRegistration register a new employee to the database
func (p *postgresDBRepo) UserPreRegistration(userType, firstName, lastName, email, mobile string) (int , error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`
		INSERT INTO %s (first_name, last_name, email, mobile, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) returning id
	`, userType)

	var id int
	err := p.DB.QueryRowContext(ctx, stmt,
		firstName,
		lastName,
		email,
		mobile,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetEmployeeDetails retrive detailed info about an employee
func (p *postgresDBRepo) GetEmployeeByID(id int) (models.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var employee models.Employee

	query := `
		SELECT 
			e.id, e.user_name, e.first_name, e.last_name, e.address, e.email, e.fb_id, e.whatsapp_id, e.x_id, e.linkedin_id, e.github_id, e.mobile, e.image_link, e.account_status_id,
			e.credits, e.task_completed, e.task_cancelled, e.rating, e.created_at, e.updated_at, es.id, es.name, es.description
		FROM
			employees e
			LEFT JOIN employee_status es on (e.account_status_id = es.id)
		WHERE
			e.id = $1
		`

	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&employee.ID,
		&employee.UserName,
		&employee.FirstName,
		&employee.LastName,
		&employee.Address,
		&employee.Email,
		&employee.FacebookID,
		&employee.WhatsappID,
		&employee.TwitterID,
		&employee.LinkedinID,
		&employee.GithubID,
		&employee.Mobile,
		&employee.ImageLink,
		&employee.AccountStatusID,
		&employee.Credits,
		&employee.TaskCompleted,
		&employee.TaskCancelled,
		&employee.Rating,
		&employee.CreatedAt,
		&employee.UpdatedAt,
		&employee.AccountStatus.ID,
		&employee.AccountStatus.Name,
		&employee.AccountStatus.Description,
	)

	return employee, err
}

// GetEmployeeListPaginated returns a chunks of employees
// if accountType == all, it will return list all employee accounts
// if accountType == active, it will return list of active employee accounts
// if accountType == ex, it will return list of all ex-employee account
// if accountType == suspended, it will return list of suspended employee's account
// if accountType == resigned, it will return list of resigned employee's account
func (p *postgresDBRepo) GetEmployeeListPaginated(accountType string, pageSize, currentPageIndex int) ([]*models.Employee, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	offset := (currentPageIndex - 1) * pageSize
	var employees []*models.Employee

	query := `
		SELECT 
			e.id, e.first_name, e.last_name, e.mobile, e.account_status_id,
			e.credits, e.rating, e.task_completed, e.created_at
		FROM
			employees e
		`

	trails := `
		 ORDER BY
			e.id asc
		LIMIT $1 OFFSET $2`
	newQuery := `
	SELECT 
		COUNT(e.id)
	FROM
		employees e
	`
	var rows *sql.Rows
	var err error

	if accountType == "all" {
		query = query + trails
	} else if accountType == "active" {
		query += ` WHERE e.account_status_id = 1` + trails
		newQuery += ` WHERE e.account_status_id = 1`
	} else if accountType == "ex" {
		query += ` WHERE e.account_status_id IN (2, 3)` + trails
		newQuery += ` WHERE e.account_status_id IN (2, 3)`
	} else if accountType == "suspended" {
		query += ` WHERE e.account_status_id = 2` + trails
		newQuery += ` WHERE e.account_status_id = 2`
	} else if accountType == "resigned" {
		query += ` WHERE e.account_status_id = 3` + trails
		newQuery += ` WHERE e.account_status_id = 3`
	} else {
		return employees, 0, errors.New("please enter correct parameter to get employees list")
	}

	rows, err = p.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return employees, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Employee
		err = rows.Scan(
			&e.ID,
			&e.FirstName,
			&e.LastName,
			&e.Mobile,
			&e.AccountStatusID,
			&e.Credits,
			&e.Rating,
			&e.TaskCompleted,
			&e.CreatedAt,
		)
		if err != nil {
			return employees, 0, err
		}
		employees = append(employees, &e)
	}

	var totalRecords int
	countRow := p.DB.QueryRowContext(ctx, newQuery)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return employees, 0, err
	}
	return employees, totalRecords, nil
}
//UpdateEmployeeAccountStatus updates the account status of an employee
func (p *postgresDBRepo) UpdateEmployeeAccountStatusByID(id, accountStatusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `
		UPDATE employees
		SET account_status_id = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := p.DB.ExecContext(ctx, stmt, accountStatusID, time.Now(), id)
	return err
}
