package app

import (
	"../database"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/keep94/weblogs"
	"log"
	"net/http"
)

var Router *mux.Router = mux.NewRouter()
var store *sessions.FilesystemStore

func Init(appConfFile string, debug bool) {

	conf := LoadConfiguration(appConfFile)
	if debug {
		conf.Debug = true
	}
	log.Print("Debug=", conf.Debug)

	if len(conf.CookieAuthKey) < 32 {
		log.Print("Random CookieAuthKey")
		conf.CookieAuthKey = string(securecookie.GenerateRandomKey(32))
	}
	if len(conf.CookieEncKey) < 32 {
		log.Print("Random CookieEncKey")
		conf.CookieEncKey = string(securecookie.GenerateRandomKey(32))
	}

	store = sessions.NewFilesystemStore(conf.SessionStoreDir,
		[]byte(conf.CookieAuthKey),
		[]byte(conf.CookieEncKey))

	database.Init(conf.Database)

	contextHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer context.Clear(r)
			h.ServeHTTP(w, r)
		})
	}

	http.Handle("/static/", http.StripPrefix("/static",
		http.FileServer(http.Dir(conf.StaticDir))))
	http.Handle("/", weblogs.Handler(contextHandler(csrfHandler(Router))))
}

func Close() {
	log.Print("Closing...")
	database.Close()
}
