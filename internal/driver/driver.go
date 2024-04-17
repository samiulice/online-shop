package driver

import (
	"database/sql"
	"time"

	
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const maxOpenDBConn = 10
const maxIdleDBConn = 10
const maxDBLifeTime = 5 * time.Minute

// ConnectSQL connects database pool for Postgres
func ConnectDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenDBConn)
	db.SetMaxIdleConns(maxIdleDBConn)
	db.SetConnMaxLifetime(maxDBLifeTime)

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
