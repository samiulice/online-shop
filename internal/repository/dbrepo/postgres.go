package dbrepo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"online_store/internal/models"
	"time"
)

// GetDate return a date package for specific id
func (m *postgresDBRepo) GetDate(id int) (models.Date, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var d models.Date

	query := `
		SELECT id, name, description, is_recurring, plan_id, plan_title, plan_description, 
			package_weight, package_price, stock_level, coalesce(image_link, ''), created_at, updated_at
		FROM dates
		WHERE id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)

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
func (m *postgresDBRepo) InsertTransaction(txn models.Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO transactions (amount, currency, payment_intent, payment_method, last_four_digits, bank_return_code, transaction_status_id, expiry_month, expiry_year, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id
	`

	var id int
	err := m.DB.QueryRowContext(ctx, stmt,
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
func (m *postgresDBRepo) InsertOrder(order models.Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO orders (dates_id, transaction_id, customer_id, status_id, quantity, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id
	`

	var id int
	err := m.DB.QueryRowContext(ctx, stmt,
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
func (m *postgresDBRepo) InsertCustomer(customer models.Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO customers (first_name, last_name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) returning id
	`

	var id int
	err := m.DB.QueryRowContext(ctx, stmt,
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

func (m *postgresDBRepo) GetOrdersHistory(statusType string) ([]*models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var orders []*models.Order

	query := `
		SELECT 
			o.id, o.dates_id, o.transaction_id, o.customer_id, o.status_id,
			o.quantity, o.amount, o.created_at, o.updated_at, 
			d.id, d.name, t.id, t.amount, t.currency, t.last_four_digits,
			t.expiry_month, t.expiry_year, t.payment_intent,
			t.bank_return_code, c.id, c.first_name, c.last_name, c.email, c.image_link
		FROM
			orders o
			LEFT JOIN dates d on (o.dates_id = d.id)
			LEFT JOIN transactions t on (o.transaction_id = t.id)
			LEFT JOIN customers c on (o.customer_id = c.id)
		`

	trails := `
		 ORDER BY
			o.created_at desc`

	var rows *sql.Rows
	var err error

	if statusType == "all" {
		query = query + trails
		rows, err = m.DB.QueryContext(ctx, query)
	} else if statusType == "completed" {
		query = query + ` WHERE o.status_id = 1` + trails
		rows, err = m.DB.QueryContext(ctx, query)
	} else if statusType == "refunded" {
		query = query + ` WHERE o.status_id = 2` + trails
		rows, err = m.DB.QueryContext(ctx, query)
	} else if statusType == "cancelled" {
		query = query + ` WHERE o.status_id = 3` + trails
		rows, err = m.DB.QueryContext(ctx, query)
	} else if statusType == "one-off" {
		query = query + ` WHERE d.is_recurring = 0` + trails
		rows, err = m.DB.QueryContext(ctx, query)
	} else if statusType == "subscriptions" {
		query = query + ` WHERE d.is_recurring = 1` + trails
		rows, err = m.DB.QueryContext(ctx, query)
	} else {
		return orders, errors.New("invalid function parameter for the database function call")
	}

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
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFourDigits,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
			&o.Customer.ImageLink,
		)
		if err != nil {
			return orders, err
		}
		orders = append(orders, &o)
	}

	// return orders, errors.New("testing errors")
	return orders, nil
}

// Database Functions that relates to User Account activity
// GetUserbyUserName gets a user by userName
func (m *postgresDBRepo) GetUserDetails(index, paramType string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//paramType : ID >>> id
	//paramType : username >>> user_name
	//paramType : (contains @) >>> email
	var u models.User

	query := `
		SELECT id, user_name, first_name, last_name, email, password, coalesce(image_link, ''), created_at, updated_at
		FROM users
		WHERE ` + paramType + ` = $1`
	row := m.DB.QueryRowContext(ctx, query, index)

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

// UpdatePasswordByUserID updates account password for a user
func (m *postgresDBRepo) UpdatePasswordByUserID(id, newPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `
		UPDATE 
			users 
		SET 
			password= $1
		WHERE id= $2`

	_, err := m.DB.ExecContext(ctx, stmt, newPassword, id)
	return err
}

// Function that relates to the Token
// InsertToken inserts token to database
func (m *postgresDBRepo) InsertToken(t *models.Token, u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Delete existing tokens for the user
	stmt := `DELETE FROM tokens WHERE user_id = $1`
	_, err := m.DB.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	stmt = `
		INSERT INTO tokens (user_id, name, email, token_hash, expiry, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = m.DB.ExecContext(ctx, stmt, //not bothering about the result
		u.ID,
		u.FirstName,
		u.Email,
		t.Hash,
		t.Expiry,
		time.Now(),
		time.Now(),
	)

	return err
}

// GetUserbyToken returns user info from tokens table
func (m *postgresDBRepo) GetUserbyToken(token string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var u models.User
	var err error

	query := `
			SELECT u.id, u.first_name, u.last_name, u.email
			FROM users u
				INNER JOIN tokens t ON (u.id = t.user_id)
			WHERE t.token_hash = $1
				AND t.expiry > $2
	`

	row := m.DB.QueryRowContext(ctx, query, tokenHash[:], time.Now())

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
