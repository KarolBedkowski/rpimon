package app

import (
	//	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app/mw"
	gzip "k.prv/rpimon/app/mw/gziphander"
	"k.prv/rpimon/app/session"
	"k.prv/rpimon/cfg"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strconv"
)

// App main router
var Router = mux.NewRouter()

// Init - Initialize application
func Init(appConfFile string, debug int) *cfg.AppConfiguration {

	Router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	conf := cfg.LoadConfiguration(appConfFile)
	if debug == 0 {
		conf.Debug = false
	} else if debug == 1 {
		conf.Debug = true
	} // other: use value from config

	l.Init(conf.LogFilename, conf.Debug)
	l.Print("Debug=", conf.Debug)

	session.InitSessionStore(conf)

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
}

// GetNamedURL - Return url for named route and parameters
func GetNamedURL(name string, pairs ...interface{}) (url string) {
	route := Router.Get(name)
	if route == nil {
		l.Error("GetNamedURL " + name + " error")
		return ""
	}
	strpairs := make([]string, len(pairs))
	for idx := 0; idx < len(pairs); idx += 2 {
		strpairs[idx] = pairs[idx].(string)
		val := pairs[idx+1]
		switch val.(type) {
		case uint64:
			strpairs[idx+1] = strconv.FormatUint(val.(uint64), 10)

		case uint:
			i := val.(uint)
			strpairs[idx+1] = strconv.FormatUint(uint64(i), 10)

		case int:
			i := val.(int)
			strpairs[idx+1] = strconv.FormatInt(int64(i), 10)

		default:
			var ok bool
			strpairs[idx+1], ok = val.(string)
			if !ok {
				l.Error("web.GetNamedURL param error param=%#v, val=%#v", strpairs[idx], val)
			}
		}
	}

	rurl, err := route.URL(strpairs...)
	if err != nil {
		l.Error("GetNamedURL " + name + " error " + err.Error())
		return ""
	}
	return rurl.String()
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	Render404(w, r)
}
