package main

import (
	"flag"
	"k.prv/rpimon/app"
	"k.prv/rpimon/modules"
	mfiles "k.prv/rpimon/modules/files"
	mmpd "k.prv/rpimon/modules/mpd"
	mnet "k.prv/rpimon/modules/network"
	mnotepad "k.prv/rpimon/modules/notepad"
	msyslogs "k.prv/rpimon/modules/system/logs"
	msystother "k.prv/rpimon/modules/system/other"
	msysproc "k.prv/rpimon/modules/system/process"
	"k.prv/rpimon/monitor"
	"k.prv/rpimon/pages/auth"
	pmain "k.prv/rpimon/pages/main"
	pstorage "k.prv/rpimon/pages/storage"
	pusers "k.prv/rpimon/pages/users"
	putils "k.prv/rpimon/pages/utils"
	"log"
	"net/http"
	// _ "net/http/pprof" // /debug/pprof/
	"runtime"
	//"time"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configFilename := flag.String("conf", "./config.json", "Configuration filename")
	debug := flag.Int("debug", -1, "Run in debug mode (1) or normal (0)")
	flag.Parse()

	conf := app.Init(*configFilename, *debug)

	if !conf.Debug {
		log.Printf("NumCPU: %d", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// cleanup
	cleanChannel := make(chan os.Signal, 1)
	signal.Notify(cleanChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-cleanChannel
		app.Close()
		modules.ShutdownModules()
		os.Exit(1)
	}()

	app.Router.HandleFunc("/", handleHome)
	auth.CreateRoutes(app.Router.PathPrefix("/auth"))
	pmain.CreateRoutes(app.Router.PathPrefix("/main"))
	pstorage.CreateRoutes(app.Router.PathPrefix("/storage"))
	putils.Init(conf.UtilsFilename)
	putils.CreateRoutes(app.Router.PathPrefix("/utils"))
	pmpd.Init(conf.MpdHost)
	pmpd.CreateRoutes(app.Router.PathPrefix("/mpd"))
	pusers.CreateRoutes(app.Router.PathPrefix("/users"))

	modules.Register(mnet.GetModule())
	modules.Register(mnet.GetNFSModule())
	modules.Register(mnet.GetSambaModule())
	modules.Register(mfiles.GetModule())
	modules.Register(mmpd.GetModule())
	modules.Register(mnotepad.GetModule())
	modules.Register(msystother.GetModule())
	modules.Register(msysproc.GetModule())
	modules.Register(msyslogs.GetModule())

	modules.InitModules(conf, app.Router)

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

	monitor.Init(conf.MonitorUpdateInterval)

	if conf.HTTPSAddress != "" {
		log.Printf("Listen: %s", conf.HTTPSAddress)
		if conf.HTTPAddress != "" {
			go func() {
				if err := http.ListenAndServeTLS(conf.HTTPSAddress,
					conf.SslCert, conf.SslKey, nil); err != nil {
					log.Fatalf("Error listening https, %v", err)
				}
			}()
		} else {
			if err := http.ListenAndServeTLS(conf.HTTPSAddress,
				conf.SslCert, conf.SslKey, nil); err != nil {
				log.Fatalf("Error listening https, %v", err)
			}
		}
	}

	if conf.HTTPAddress != "" {
		log.Printf("Listen: %s", conf.HTTPAddress)
		if err := http.ListenAndServe(conf.HTTPAddress, nil); err != nil {
			log.Fatalf("Error listening http, %v", err)
		}
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/main/", http.StatusFound)
}
