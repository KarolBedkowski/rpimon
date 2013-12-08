package database

import (
	"encoding/csv"
	"io"
	l "k.prv/rpimon/helpers/logging"
	"os"
	"strings"
)

var users = make(map[string]*User, 5)

func Init(usersFile string, debug bool) {
	LoadUsers(usersFile)
}

func LoadUsers(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		l.Error("Error:", err)
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			l.Error("Error:", err)
			return err
		}
		user := &User{Login: record[0],
			Password: record[1],
			Privs:    strings.Split(record[2], " ")}
		users[user.Login] = user
	}
	return nil
}

func SelectOneUser(login string) *User {
	return users[login]
}
