package database

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB(connStr string) {
	var err error
	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}
	createTable()
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id SERIAL PRIMARY KEY,
		service_name TEXT,
		price INT,
		user_id TEXT,
		start_date DATE,
		end_date DATE
	)`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
