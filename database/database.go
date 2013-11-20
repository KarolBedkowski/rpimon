package database

import (
	"../helpers"
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var Database gorp.DbMap

func Init(dbFileName string) {
	db, err := sql.Open("sqlite3", dbFileName)
	helpers.CheckErr(err, "sql.Open failed")
	Database.Db = db
	Database.Dialect = gorp.SqliteDialect{}

	// Register objects
	Database.AddTableWithName(User{}, "users").SetKeys(true, "Id")

	// Create tables
	err = Database.CreateTablesIfNotExists()
	helpers.CheckErr(err, "Create tables failed")

	// Bootstrap
	BootstrapUsers()
}

func Close() {
	Database.Db.Close()
}
