package main

import (
	"flag"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	mauth "k.prv/rpimon/modules/auth"
	mfiles "k.prv/rpimon/modules/files"
	mmain "k.prv/rpimon/modules/main"
	mmonitor "k.prv/rpimon/modules/monitor"
	mmpd "k.prv/rpimon/modules/mpd"
	mnet "k.prv/rpimon/modules/network"
	mnotepad "k.prv/rpimon/modules/notepad"
	mpref "k.prv/rpimon/modules/preferences"
	mstorage "k.prv/rpimon/modules/storage"
	msmart "k.prv/rpimon/modules/storage/smart"
	msystem "k.prv/rpimon/modules/system"
	msyslogs "k.prv/rpimon/modules/system/logs"
	msystother "k.prv/rpimon/modules/system/other"
	msysproc "k.prv/rpimon/modules/system/process"
	msysusers "k.prv/rpimon/modules/system/users"
	mutls "k.prv/rpimon/modules/utils"
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
		context.ShutdownModules()
		os.Exit(1)
	}()

	app.Router.HandleFunc("/", handleHome)
	context.RegisterModule(mauth.Module)
	context.RegisterModule(mmain.Module)
	context.RegisterModule(mpref.Module)
	context.RegisterModule(mfiles.Module)
	context.RegisterModule(mmpd.Module)
	context.RegisterModule(mnet.Module)
	context.RegisterModule(mnet.NFSModule)
	context.RegisterModule(mnet.SambaModule)
	context.RegisterModule(mnotepad.Module)
	context.RegisterModule(mstorage.Module)
	context.RegisterModule(msmart.Module)
	context.RegisterModule(msyslogs.Module)
	context.RegisterModule(msystother.Module)
	context.RegisterModule(msysproc.Module)
	context.RegisterModule(msysusers.Module)
	context.RegisterModule(mutls.Module)
	context.RegisterModule(msystem.Module)
	context.RegisterModule(mmonitor.Module)
	context.InitModules(conf, app.Router)

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
	http.Redirect(w, r, app.GetNamedURL("main-index"), http.StatusFound)
}
