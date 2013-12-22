package main

import (
	"flag"
	"k.prv/rpimon/app"
	"k.prv/rpimon/pages/auth"
	pfiles "k.prv/rpimon/pages/files"
	plogs "k.prv/rpimon/pages/logs"
	pmain "k.prv/rpimon/pages/main"
	pmpd "k.prv/rpimon/pages/mpd"
	pnet "k.prv/rpimon/pages/net"
	pproc "k.prv/rpimon/pages/process"
	pstorage "k.prv/rpimon/pages/storage"
	pusers "k.prv/rpimon/pages/users"
	putils "k.prv/rpimon/pages/utils"
	"log"
	"net/http"
	// _ "net/http/pprof" // /debug/pprof/
	//"time"
)

func main() {
	configFilename := flag.String("conf", "./config.json", "Configuration filename")
	debug := flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()

	conf := app.Init(*configFilename, *debug)
	defer app.Close()

	app.Router.HandleFunc("/", handleHome)
	auth.CreateRoutes(app.Router.PathPrefix("/auth"))
	pmain.CreateRoutes(app.Router.PathPrefix("/main"))
	pnet.CreateRoutes(app.Router.PathPrefix("/net"))
	pstorage.CreateRoutes(app.Router.PathPrefix("/storage"))
	putils.Init(conf.UtilsFilename)
	putils.CreateRoutes(app.Router.PathPrefix("/utils"))
	pmpd.Init(conf.MpdHost)
	pmpd.CreateRoutes(app.Router.PathPrefix("/mpd"))
	plogs.CreateRoutes(app.Router.PathPrefix("/logs"))
	pusers.CreateRoutes(app.Router.PathPrefix("/users"))
	pproc.CreateRoutes(app.Router.PathPrefix("/process"))
	pfiles.Init(conf.BrowserConf)
	pfiles.CreateRoutes(app.Router.PathPrefix("/files"))

	/* for filesystem store
	go app.ClearSessionStore()
	// clear session task
	ticker := time.NewTicker(time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				app.ClearSessionStore()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	*/
	if conf.HttpsAddress != "" {
		log.Printf("Listen: %s", conf.HttpsAddress)
		if conf.HttpAddress != "" {
			go func() {
				if err := http.ListenAndServeTLS(conf.HttpsAddress,
					conf.SslCert, conf.SslKey, nil); err != nil {
					log.Fatalf("Error listening https, %v", err)
				}
			}()
		} else {
			if err := http.ListenAndServeTLS(conf.HttpsAddress,
				conf.SslCert, conf.SslKey, nil); err != nil {
				log.Fatalf("Error listening https, %v", err)
			}
		}
	}

	if conf.HttpAddress != "" {
		log.Printf("Listen: %s", conf.HttpAddress)
		if err := http.ListenAndServe(conf.HttpAddress, nil); err != nil {
			log.Fatalf("Error listening http, %v", err)
		}
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/main/", http.StatusFound)
}
