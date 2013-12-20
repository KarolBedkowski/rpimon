package app

import (
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"k.prv/rpimon/database"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	nurl "net/url"
)

// App main router
var Router = mux.NewRouter()

// Init - Initialize application
func Init(appConfFile string, debug bool) *AppConfiguration {

	conf := LoadConfiguration(appConfFile)
	if debug {
		conf.Debug = true
	}
	l.Init(conf.LogFilename, conf.Debug)

	l.Print("Debug=", conf.Debug)

	initSessionStore(conf)
	database.Init(conf.Users, conf.Debug)

	http.Handle("/static/", http.StripPrefix("/static",
		http.FileServer(http.Dir(conf.StaticDir))))
	http.Handle("/", logHandler(contextHandler(csrfHandler(Router))))
	return conf
}

// Close application
func Close() {
	l.Info("Closing...")
}

// GetNamedURL - Return url for named route and parameters
func GetNamedURL(name string, pairs ...string) (url string) {
	url = ""
	rurl, err := Router.Get(name).URL(pairs...)
	if err != nil {
		l.Warn("GetNamedURL error %s", err)
		return
	}
	url = rurl.String()
	return
}

func PairsToQuery(pairs ...string) (query string) {
	query = ""
	pairsLen := len(pairs)
	if pairsLen == 0 {
		return
	}
	if pairsLen%2 != 0 {
		l.Warn("GetNamedURL error - wron number of argiments")
		return
	}
	query += "?"
	for idx := 0; idx < pairsLen; idx += 2 {
		query += pairs[idx] + "=" + nurl.QueryEscape(pairs[idx+1])
	}
	return
}
