package helpers

import (
	"k.prv/rpimon/logging"
	"net/http"
	nurl "net/url"
	"os"
	"strconv"
	"strings"
	"unicode"
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
func BuildQuery(pairs ...interface{}) (query string) {
	pairsLen := len(pairs)
	if pairsLen == 0 {
		return ""
	}
	if pairsLen%2 != 0 {
		logging.Error("helpers.BuildQuery error - wrong number of arguments: %v", pairs)
		return ""
	}
	query = "?"
	for idx := 0; idx < pairsLen; idx += 2 {
		name := pairs[idx].(string)
		val := pairs[idx+1]
		var valstr string
		switch val.(type) {
		case int:
			valstr = strconv.Itoa(val.(int))
		case uint64:
			valstr = strconv.FormatUint(val.(uint64), 10)
		default:
			valstr = val.(string)
		}
		query += name + "=" + nurl.QueryEscape(valstr)
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

// StringToArgs convert given string to command and arguments
func StringToArgs(inp string) (cmd string, args []string) {
	// https://groups.google.com/d/msg/golang-nuts/pNwqLyfl2co/APaZSSvQUAAJ
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)

		}
	}

	fields := strings.FieldsFunc(inp, f)
	if len(fields) < 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}
