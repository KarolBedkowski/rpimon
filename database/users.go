package database

import (
//	"k.prv/rpimon/helpers"
//	l "k.prv/rpimon/helpers/logging"
)

type User struct {
	Login    string
	Password string
	Privs    []string
}

func GetUserByLogin(login string) (user *User) {
	user = SelectOneUser(login)
	return user
}
