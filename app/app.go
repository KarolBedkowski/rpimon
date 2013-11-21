package app

import (
	"../database"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/keep94/weblogs"
	"log"
	"net/http"
	nurl "net/url"
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

func GetNamedUrl(name string, pairs ...string) (url string, err error) {
	url = ""
	rurl, err := Router.Get(name).URL()
	if err != nil {
		return
	}
	url = rurl.String()
	pairs_len := len(pairs)
	if pairs_len == 0 {
		return
	}
	if pairs_len%2 != 0 {
		err = fmt.Errorf("Requred pairs of arguments")
		return
	}
	url += "?"
	for idx := 0; idx < pairs_len; idx += 2 {
		url += pairs[idx] + "=" + nurl.QueryEscape(pairs[idx+1])
	}
	return
}
