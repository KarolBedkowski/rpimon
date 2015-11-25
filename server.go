package main

import (
	"flag"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/model"

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
	msyshw "k.prv/rpimon/modules/system/hw"
	msyslogs "k.prv/rpimon/modules/system/logs"
	msysproc "k.prv/rpimon/modules/system/process"
	msysusers "k.prv/rpimon/modules/system/users"
	mutls "k.prv/rpimon/modules/utils"
	mworker "k.prv/rpimon/modules/worker"
	"k.prv/rpimon/resources"
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
	log.Printf("Starting... ver %s", context.AppVersion)
	configFilename := flag.String("conf", "./config.json", "Configuration filename")
	debug := flag.Int("debug", -1, "Run in debug mode (1) or normal (0)")
	forceLocalFiles := flag.Bool("forceLocalFiles", false, "Force use local files instead of embended assets")
	localFilesPath := flag.String("localFilesPath", ".", "Path to static and templates directory")
	flag.Parse()

	conf := app.Init(*configFilename, *debug)
	model.Open(conf.Database)

	if !conf.Debug {
		log.Printf("NumCPU: %d", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if resources.Init(*forceLocalFiles, *localFilesPath) {
		log.Printf("Using embended resources...")
	} else {
		log.Printf("Using local files...")
	}

	defer func() {
		if e := recover(); e != nil {
			model.Close()
		}
	}()

	// cleanup
	cleanChannel := make(chan os.Signal, 1)
	signal.Notify(cleanChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-cleanChannel
		app.Close()
		model.Close()
		context.ShutdownModules()
		os.Exit(0)
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
	context.RegisterModule(msyshw.Module)
	context.RegisterModule(msysproc.Module)
	context.RegisterModule(msysusers.Module)
	context.RegisterModule(mutls.Module)
	context.RegisterModule(msystem.Module)
	context.RegisterModule(mmonitor.Module)
	context.RegisterModule(mworker.Module)
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
		go func() {
			if err := http.ListenAndServeTLS(conf.HTTPSAddress,
				conf.SslCert, conf.SslKey, nil); err != nil {
				log.Fatalf("Error listening https, %v", err)
			}
		}()
	}

	if conf.HTTPAddress != "" {
		log.Printf("Listen: %s", conf.HTTPAddress)
		go func() {
			if err := http.ListenAndServe(conf.HTTPAddress, nil); err != nil {
				log.Fatalf("Error listening http, %v", err)
			}
		}()
	}
	done := make(chan bool)
	<-done
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, app.GetNamedURL("main-index"), http.StatusFound)
}
