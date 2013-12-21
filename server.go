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
	_ "net/http/pprof"
	//"time"
)

func main() {
	configFilename := flag.String("conf", "./config.json", "Configuration filename")
	debug := flag.Bool("debug", false, "Run in debug mode")
	httpAddr := flag.String("addr", ":8000", "HTTP server address")
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

	log.Printf("Listen: %s", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatalf("Error listening, %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/main/", http.StatusFound)
}
