package helpers

import (
	"./logging"
	"code.google.com/p/go.crypto/bcrypt"
	"os"
)

const BCRYPT_COST = 12
const MIN_PASSWORD_LENGTH = 8

func CheckErr(err error, msg string) {
	if err != nil {
		logging.Error(msg)
	}
}

func CheckErrAndDie(err error, msg string) {
	if err != nil {
		logging.Error(msg)
		os.Exit(1)
	}
}

func ComparePassword(user_password string, candidate_password string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(user_password), []byte(candidate_password))
	return
}

func CreatePassword(password string) string {
	password_hashed, err := bcrypt.GenerateFromPassword([]byte(password), BCRYPT_COST)
	if err != nil {
		panic(err)
	}
	return string(password_hashed)
}
