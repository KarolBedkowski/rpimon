package app

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"k.prv/rpimon/database"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	nurl "net/url"
	"os"
)

// App main router
var Router = mux.NewRouter()
var store *sessions.FilesystemStore

// Init - Initialize application
func Init(appConfFile string, debug bool) *AppConfiguration {

	conf := LoadConfiguration(appConfFile)
	if debug {
		conf.Debug = true
	}
	l.Init(conf.LogFilename, conf.Debug)

	l.Print("Debug=", conf.Debug)

	if len(conf.CookieAuthKey) < 32 {
		l.Info("Random CookieAuthKey")
		conf.CookieAuthKey = string(securecookie.GenerateRandomKey(32))
	}
	if len(conf.CookieEncKey) < 32 {
		l.Info("Random CookieEncKey")
		conf.CookieEncKey = string(securecookie.GenerateRandomKey(32))
	}

	err := os.MkdirAll(conf.SessionStoreDir, os.ModeDir)
	if err != nil && !os.IsExist(err) {
		l.Error("Createing dir for session store failed ", err)
	}

	store = sessions.NewFilesystemStore(conf.SessionStoreDir,
		[]byte(conf.CookieAuthKey),
		[]byte(conf.CookieEncKey))

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
	rurl, err := Router.Get(name).URL()
	if err != nil {
		l.Warn("GetNamedURL error %s", err)
		return
	}
	url = rurl.String()
	pairsLen := len(pairs)
	if pairsLen == 0 {
		return
	}
	if pairsLen%2 != 0 {
		l.Warn("GetNamedURL error - wron number of argiments")
		return
	}
	url += "?"
	for idx := 0; idx < pairsLen; idx += 2 {
		url += pairs[idx] + "=" + nurl.QueryEscape(pairs[idx+1])
	}
	return
}
