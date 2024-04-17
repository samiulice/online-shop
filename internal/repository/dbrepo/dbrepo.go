package dbrepo

import (
	"database/sql"
	"online_store/internal/repository"
)

type postgresDBRepo struct {
	DB  *sql.DB
}

func NewDBRepo(conn *sql.DB) repository.DatabaseRepo {
	return &postgresDBRepo{
		DB:  conn,
	}
}