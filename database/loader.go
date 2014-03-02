package database

import (
	"encoding/json"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
)

// UserDB - virtual database
type UserDB struct {
	Users []*User
}

var database UserDB

var AllPrivs = []string{
	"admin", "mpd", "files", "notepad",
}

var dbfilename string

// Init structures
func Init(filename string, debug bool) {
	l.Info("UserDB.Init from: %s", filename)
	dbfilename = filename
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("UserDB.Init Error:", err)
		createDummyDatabase()
		return
	}
	err = json.Unmarshal(file, &database)
	if err != nil {
		l.Error("UserDB.Init Error: %s", err.Error())
		createDummyDatabase()
	}
	l.Info("UserDB.Init loaded users: %d", len(database.Users))
	return
}

func Save() error {
	l.Printf("UserDB.Save %s\n", dbfilename)
	data, err := json.Marshal(database)
	if err != nil {
		l.Printf("UserDB.Save: error marshal configuration: %s\n", err)
		return err
	}
	return ioutil.WriteFile(dbfilename, data, 0)
}

func createDummyDatabase() {
	l.Info("Creating default user 'user', 'guest', 'admin'")
	//Create fake user
	database.Users = []*User{
		&User{
			Login:    "guest",
			Password: "",
			Privs:    nil,
		},
		&User{
			Login:    "user",
			Password: "",
			Privs:    []string{"mpd", "files"},
		},
		&User{
			Login:    "admin",
			Password: "",
			Privs:    []string{"admin", "mpd", "files"},
		},
	}
}
