package app

import (
	"../database"
	"github.com/coopernurse/gorp"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/keep94/weblogs"
	"log"
	"net/http"
)

type webApp struct {
	Router   *mux.Router
	store    *sessions.FilesystemStore
	Database *gorp.DbMap
}

// Current Web application
var App webApp

func NewWebApp(appConfFile string, debug bool) *webApp {

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

	App.Router = mux.NewRouter()
	App.store = sessions.NewFilesystemStore(conf.SessionStoreDir,
		[]byte(conf.CookieAuthKey),
		[]byte(conf.CookieEncKey))
	App.Database = database.Init(conf.Database)

	contextHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer context.Clear(r)
			h.ServeHTTP(w, r)
		})
	}

	http.Handle("/static/", http.StripPrefix("/static",
		http.FileServer(http.Dir(conf.StaticDir))))
	http.Handle("/", weblogs.Handler(contextHandler(csrfHandler(App.Router))))

	return &App
}

func (app *webApp) Close() {
	log.Print("Closing...")
	app.Database.Db.Close()
}
