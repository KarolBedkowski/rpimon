package main

import (
	"flag"
	"k.prv/rpimon/app"
	mfiles "k.prv/rpimon/modules/files"
	mmpd "k.prv/rpimon/modules/mpd"
	mnet "k.prv/rpimon/modules/network"
	mnotepad "k.prv/rpimon/modules/notepad"
	mstorage "k.prv/rpimon/modules/storage"
	msmart "k.prv/rpimon/modules/storage/smart"
	msystem "k.prv/rpimon/modules/system"
	msyslogs "k.prv/rpimon/modules/system/logs"
	msystother "k.prv/rpimon/modules/system/other"
	msysproc "k.prv/rpimon/modules/system/process"
	msysusers "k.prv/rpimon/modules/system/users"
	mutls "k.prv/rpimon/modules/utils"
	"k.prv/rpimon/monitor"
	"k.prv/rpimon/pages/auth"
	pmain "k.prv/rpimon/pages/main"
	pmodules "k.prv/rpimon/pages/preferences/modules"
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
		app.ShutdownModules()
		os.Exit(1)
	}()

	app.Router.HandleFunc("/", handleHome)
	auth.CreateRoutes(app.Router.PathPrefix("/auth"))
	pmain.CreateRoutes(app.Router.PathPrefix("/main"))
	pmodules.CreateRoutes(app.Router.PathPrefix("/pref/modules"))

	app.RegisterModule(mnet.Module)
	app.RegisterModule(mnet.NFSModule)
	app.RegisterModule(mnet.SambaModule)
	app.RegisterModule(mfiles.Module)
	app.RegisterModule(mmpd.Module)
	app.RegisterModule(mnotepad.Module)
	app.RegisterModule(msystem.Module)
	app.RegisterModule(msystother.Module)
	app.RegisterModule(msysproc.Module)
	app.RegisterModule(msyslogs.Module)
	app.RegisterModule(msysusers.Module)
	app.RegisterModule(mstorage.Module)
	app.RegisterModule(msmart.Module)
	app.RegisterModule(mutls.Module)

	app.InitModules(conf, app.Router)

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
