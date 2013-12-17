package app

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"k.prv/rpimon/database"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	nurl "net/url"
	"os"
	"path/filepath"
	"time"
)

var Router *mux.Router = mux.NewRouter()
var store *sessions.FilesystemStore

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

	store = sessions.NewFilesystemStore(conf.SessionStoreDir,
		[]byte(conf.CookieAuthKey),
		[]byte(conf.CookieEncKey))

	database.Init(conf.Users, conf.Debug)

	http.Handle("/static/", http.StripPrefix("/static",
		http.FileServer(http.Dir(conf.StaticDir))))
	http.Handle("/", logHandler(contextHandler(csrfHandler(Router))))
	return conf
}

func Close() {
	l.Info("Closing...")
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

func ClearSessionStore() error {
	l.Info("Start ClearSessionStore")
	now := time.Now()
	now = now.Add(time.Duration(-24) * time.Hour)
	filepath.Walk(Configuration.SessionStoreDir, func(path string, info os.FileInfo, err error) error {
		if now.After(info.ModTime()) {
			l.Info("Delete ", path)
			os.Remove(path)
		}
		return nil
	})
	return nil
}
