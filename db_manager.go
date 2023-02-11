package gocrudify

import "github.com/jmoiron/sqlx"

var db *sqlx.DB

func initializeDB(database *sqlx.DB) {
	if db == nil {
		db = database
	}
}
