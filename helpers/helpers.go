package helpers

import (
	"errors"
	"k.prv/rpimon/helpers/logging"
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
	if user_password == candidate_password {
		return nil
	}
	return errors.New("Wrong password")
}
