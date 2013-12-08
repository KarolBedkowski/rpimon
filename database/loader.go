package database

import (
	"encoding/json"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
)

type UsersConfig struct {
	Users []User
}

var users UsersConfig

func Init(usersFile string, debug bool) {
	LoadUsers(usersFile)
}

func LoadUsers(filename string) error {
	l.Info("LoadUsers from: %s", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("LoadUsers Error:", err)
		return err
	}
	err = json.Unmarshal(file, &users)
	if err != nil {
		l.Error("LoadUsers Error:", err)
	}
	l.Info("LoadUsers loaded: %d", len(users.Users))
	return err
}

func SelectOneUser(login string) *User {
	for _, user := range users.Users {
		if user.Login == login {
			return &user
		}
	}
	return nil
}
