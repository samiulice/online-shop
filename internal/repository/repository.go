package repository

import "online_store/internal/models"

type DatabaseRepo interface {
	GetDate(id int) (models.Date, error)
}