package repository

import (
	"online_store/internal/models"
)

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

	//Employee
	
	GetEmployeeByID(id int) (models.Employee, error)
	GetEmployeeListPaginated(accountType string, pageSize, currentPageIndex int) ([]*models.Employee, int, error)
	UpdateEmployeeAccountStatusByID(id, accountStatusID int) error
	
	//User
	GetUserDetails(index, paramType, account_type string) (models.User, error)
	UpdateUserPasswordByID(userType, id, newPassword string) error
	IsRegistered(userType, paramType, paramValue string) (int, error)
	VerifyUser(userType, searchParam, paramValue string) (int, string, string, error)
	UserPreRegistration(userType, firstName, lastName, email, mobile string) (int , error)
	//Token
	GetUserInitialData(userName, param, tableName string)(models.UserInitialData, error)
	GetAccountTypeByToken(token string) (string, error)
	InsertToken(t *models.Token, u models.UserInitialData) error
	GetUserbyToken(token string) (*models.User, error)


	//Customer
	GetCustomerProfile(index string) ([]*models.Customer, error)
}
