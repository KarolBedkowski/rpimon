package database

import (
	"encoding/json"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
)

// Database - virtual database
type Database struct {
	Users []User
}

var database Database

// Init structures
func Init(filename string, debug bool) {
	l.Info("Database.Init from: %s", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("Database.Init Error:", err)
		return
	}
	err = json.Unmarshal(file, &database)
	if err != nil {
		l.Error("Database.Init Error:", err)
	}
	l.Info("Database.Init loaded users: %d", len(database.Users))
	return
}
