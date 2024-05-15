package repository

import "online_store/internal/models"

type DatabaseRepo interface {
	GetDate(id int) (models.Date, error)
	InsertTransaction(tnx models.Transaction) (int, error)
	InsertOrder(order models.Order) (int, error)

	//Customer
	InsertCustomer(customer models.Customer) (int, error)
	//TODO:
	// GetCustomerDetails(id string) ([]*models.Customer, error)

	//Order
	GetOrdersHistory(statusType string) ([]*models.Order, error)
	GetOrdersHistoryPaginated(statusType string, pageSize, currentPageIndex int) ([]*models.Order, int, error)
	UpdateOrderStatusID(id, statusID int) error

	//Transaction
	GetTransactionsHistory(statusType string) ([]*models.Transaction, error)
	GetTransactionsHistoryPaginated(statusType string, pageSize, currentPageIndex int) ([]*models.Transaction, int, error)
	UpdateTransactionStatusID(id, statusID int) error

	//User
	GetUserDetails(index, paramType string) (models.User, error)
	UpdatePasswordByUserID(id, newPassword string) error

	//Token
	InsertToken(t *models.Token, u models.User) error
	GetUserbyToken(token string) (*models.User, error)

	//Customer
	GetCustomerProfile(index string) ([]*models.Customer, error)
}
