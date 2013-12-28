package helpers

import (
	"k.prv/rpimon/helpers/logging"
	nurl "net/url"
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

// BuildQuery format url query part from pairs key, val
func BuildQuery(pairs ...string) (query string) {
	query = ""
	pairsLen := len(pairs)
	if pairsLen == 0 {
		return
	}
	if pairsLen%2 != 0 {
		logging.Warn("GetNamedURL error - wron number of argiments")
		return
	}
	query += "?"
	for idx := 0; idx < pairsLen; idx += 2 {
		query += pairs[idx] + "=" + nurl.QueryEscape(pairs[idx+1])
	}
	return
}
