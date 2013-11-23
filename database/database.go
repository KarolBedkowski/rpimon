package database

import (
	"../helpers"
	l "../helpers/logging"
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var Database gorp.DbMap

func Init(dbFileName string, debug bool) {
	if debug {
		Database.TraceOn("[gorp]", l.Logger)
	}

	db, err := sql.Open("sqlite3", dbFileName)
	helpers.CheckErrAndDie(err, "sql.Open failed")
	Database.Db = db
	Database.Dialect = gorp.SqliteDialect{}

	// Register objects
	Database.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	Database.AddTableWithName(Profile{}, "profiles").SetKeys(true, "Id")
	Database.AddTableWithName(Privilage{}, "privilages").SetKeys(true, "Id")
	Database.AddTableWithName(UserProfile{}, "user_profile")
	Database.AddTableWithName(ProfilePrivilage{}, "profile_privilages")

	// Create tables
	err = Database.CreateTablesIfNotExists()
	helpers.CheckErrAndDie(err, "Create tables failed")

	// Bootstrap
	BootstrapUsers()
}

func Close() {
	Database.Db.Close()
}
