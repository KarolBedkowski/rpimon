package helpers

import (
	"k.prv/rpimon/helpers/logging"
	"net/http"
	nurl "net/url"
	"os"
)

// CheckErr - when err != nil log message
func CheckErr(err error, msg string) error {
	if err != nil {
		logging.Error(msg)
	}
	return err
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
	pairsLen := len(pairs)
	if pairsLen == 0 {
		return ""
	}
	if pairsLen%2 != 0 {
		logging.Warn("helpers.BuildQuery error - wron number of argiments: %v", pairs)
		return ""
	}
	query = "?"
	for idx := 0; idx < pairsLen; idx += 2 {
		query += pairs[idx] + "=" + nurl.QueryEscape(pairs[idx+1])
	}
	return
}

// GetParam return form value and ok=true when param is in request and != "" or generate http.Error
func GetParam(w http.ResponseWriter, r *http.Request, param string) (value string, ok bool) {
	var paramL []string
	if paramL, ok = r.Form[param]; ok {
		value = paramL[0]
		if value != "" {
			ok = true
		}
	}
	if !ok {
		http.Error(w, "missing id", http.StatusBadRequest)
	}
	return
}

// CheckValueInStrList return true when value is in list
func CheckValueInStrList(list []string, value string) (inlist bool) {
	for _, val := range list {
		if value == val {
			return true
		}
	}
	return
}

// CheckValueInDictOfList return true when value exists in any list in map
func CheckValueInDictOfList(dict map[string][]string, value string) (inlist bool) {
	for _, list := range dict {
		for _, val := range list {
			if value == val {
				return true
			}
		}
	}
	return
}
