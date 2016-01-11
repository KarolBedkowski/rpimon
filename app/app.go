package app

import (
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"k.prv/rpimon/cfg"
	l "k.prv/rpimon/logging"
	"net/http"
	"strconv"
)

// FORMCSRFTOKEN is csrf tokens name in forms
const FORMCSRFTOKEN = "_CsrfToken"

// App main router
var router = mux.NewRouter()

// Init - Initialize application
func Init(appConfFile string, debug int) *cfg.AppConfiguration {
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	conf := cfg.LoadConfiguration(appConfFile)
	if debug == 0 {
		conf.Debug = false
	} else if debug == 1 {
		conf.Debug = true
	} // other: use value from config

	l.Init(conf.LogFilename, conf.Debug)
	l.Info("Debug=%s", conf.Debug)

	initSessionStore(conf)

	http.Handle("/metrics", prometheus.Handler())

	router.HandleFunc("/", handleHome)
	http.Handle("/static/", prometheus.InstrumentHandler("rpimon-static", http.StripPrefix("/static",
		FileServer(http.Dir(conf.StaticDir), !conf.Debug))))
	http.Handle("/external/", prometheus.InstrumentHandler("rpimon-ext", http.StripPrefix("/external",
		http.FileServer(http.Dir("external")))))
	http.Handle("/favicon.ico", FileServer(http.Dir(conf.StaticDir), !conf.Debug))
	//context.ClearHandler()
	CSRF := csrf.Protect([]byte(conf.CSRFKey), csrf.Secure(false))
	http.Handle("/", prometheus.InstrumentHandler("rpimon", logHandler(CSRF(SessionHandler(router)))))
	return conf
}

// Close application
func Close() {
	l.Info("Closing...")
}

// GetNamedURL - Return url for named route and parameters
func GetNamedURL(name string, pairs ...interface{}) (url string) {
	route := router.Get(name)
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

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, GetNamedURL("main-index"), http.StatusFound)
}
