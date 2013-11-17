package database

import (
	"../helpers"
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var Database *gorp.DbMap

func Init() *gorp.DbMap {
	db, err := sql.Open("sqlite3", "database.sqlite")
	helpers.CheckErr(err, "sql.Open failed")
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")

	// Create tables
	err = dbmap.CreateTablesIfNotExists()
	helpers.CheckErr(err, "Create tables failed")

	Database = dbmap

	BootstrapUsers()

	return dbmap
}
