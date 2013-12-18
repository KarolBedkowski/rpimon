package helpers

import (
	"k.prv/rpimon/helpers/logging"
	"os"
)

// CheckErr - when err != nil log message
func CheckErr(err error, msg string) {
	if err != nil {
		logging.Error(msg)
	}
}

// CheckErrAndDie - when err != nil, log message and die
func CheckErrAndDie(err error, msg string) {
	if err != nil {
		logging.Error(msg)
		os.Exit(1)
	}
}
