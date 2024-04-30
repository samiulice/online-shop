package dbrepo

import (
	"context"
	"crypto/sha256"
	"online_store/internal/models"
	"strings"
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

// GetUserbyUserName gets a user by userName
func (m *postgresDBRepo) GetUserbyUserName(userName string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	userName = strings.ToLower(userName)

	index := "user_name"
	if strings.Contains(userName, "@") {
		index = "email"
	}
	var u models.User

	query := `
		SELECT id, user_name, first_name, last_name, email, password, coalesce(image_link, ''), created_at, updated_at
		FROM users
		WHERE ` + index + ` = $1`
	row := m.DB.QueryRowContext(ctx, query, userName)

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
