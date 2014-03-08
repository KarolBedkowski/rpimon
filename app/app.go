package app

import (
	//	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app/cfg"
	"k.prv/rpimon/app/mw"
	gzip "k.prv/rpimon/app/mw/gziphander"
	"k.prv/rpimon/app/session"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

// App main router
var Router = mux.NewRouter()

// Init - Initialize application
func Init(appConfFile string, debug int) *cfg.AppConfiguration {

	conf := cfg.LoadConfiguration(appConfFile)
	if debug == 0 {
		conf.Debug = false
	} else if debug == 1 {
		conf.Debug = true
	} // other: use value from config

	l.Init(conf.LogFilename, conf.Debug)
	l.Print("Debug=", conf.Debug)

	session.InitSessionStore(conf)
	cfg.InitUsers(conf.Users, conf.Debug)

	http.Handle("/static/", http.StripPrefix("/static",
		gzip.FileServer(http.Dir(conf.StaticDir), !conf.Debug)))
	http.Handle("/favicon.ico", gzip.FileServer(http.Dir(conf.StaticDir), !conf.Debug))
	//context.ClearHandler()
	http.Handle("/", mw.LogHandler(mw.CsrfHandler(session.SessionHandler(Router))))
	return conf
}

// Close application
func Close() {
	l.Info("Closing...")
	cfg.CloseConf()
}

// GetNamedURL - Return url for named route and parameters
func GetNamedURL(name string, pairs ...string) (url string) {
	route := Router.Get(name)
	if route == nil {
		l.Error("GetNamedURL %s error", name)
		return ""
	}
	rurl, err := route.URL(pairs...)
	if err != nil {
		l.Error("GetNamedURL %s error %s", name, err)
		return ""
	}
	return rurl.String()
}
