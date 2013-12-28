package app

import (
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"k.prv/rpimon/database"
	l "k.prv/rpimon/helpers/logging"
	gzip "k.prv/rpimon/lib/gziphander"
	"net/http"
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
		gzip.FileServer(http.Dir(conf.StaticDir))))
	//http.FileServer(http.Dir(conf.StaticDir))))
	http.Handle("/", logHandler(csrfHandler(context.ClearHandler(Router))))
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
