package helpers

import (
	"k.prv/rpimon/helpers/logging"
	"os"
)

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
