package dbrepo

import (
	"context"
	"online_store/internal/models"
	"time"
)

// GetDate return a date package for specific id
func (m *postgresDBRepo) GetDate(id int) (models.Date, error) {
	cntx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var d models.Date

	query := `
		SELECT 
			id, name, description, package_size, package_weight, package_price, stock_level, coalesce(image, ''), created_at, updated_at
		FROM 
			dates
		WHERE id = $1`
	row := m.DB.QueryRowContext(cntx, query, id)

	err := row.Scan(
		&d.ID,
		&d.Name,
		&d.Description,
		&d.PackageSize,
		&d.PackageWeight,
		&d.PackagePrice,
		&d.StockLevel,
		&d.Image,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	return d, err
}

// InsertTransaction inserts new transaction to the database and returns its id
func (m *postgresDBRepo) InsertTransaction(txn models.Transaction) (int, error) {
	cntx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO transactions (amount, currency, payment_intent, payment_method, last_four_digits, bank_return_code, transaction_status_id, expiry_month, expiry_year, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id
	`

	var id int
	err := m.DB.QueryRowContext(cntx, stmt,
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
	cntx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO orders (dates_id, transaction_id, customer_id, status_id, quantity, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id
	`

	var id int
	err := m.DB.QueryRowContext(cntx, stmt,
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
	cntx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO customers (first_name, last_name, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5) returning id
	`

	var id int
	err := m.DB.QueryRowContext(cntx, stmt,
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
