package main

import (
	"flag"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/logging"
	"k.prv/rpimon/model"
	mauth "k.prv/rpimon/modules/auth"
	mfiles "k.prv/rpimon/modules/files"
	mmain "k.prv/rpimon/modules/main"
	mmonitor "k.prv/rpimon/modules/monitor"
	mmpd "k.prv/rpimon/modules/mpd"
	mnet "k.prv/rpimon/modules/network"
	mnotepad "k.prv/rpimon/modules/notepad"
	mpref "k.prv/rpimon/modules/preferences"
	msensors "k.prv/rpimon/modules/sensors"
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
	log.Printf("Starting... ver %s", app.AppVersion)
	configFilename := flag.String("conf", "./config.json", "Configuration filename")
	debug := flag.Int("debug", -1, "Run in debug mode (1) or normal (0)")
	forceLocalFiles := flag.Bool("forceLocalFiles", false, "Force use local files instead of embended assets")
	localFilesPath := flag.String("localFilesPath", ".", "Path to static and templates directory")
	flag.Parse()

	conf := app.Init(*configFilename, *debug)
	model.Open(conf.Database)

	if !conf.Debug {
		l.Info("NumCPU: %d", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if resources.Init(*forceLocalFiles, *localFilesPath) {
		l.Info("Using embended resources...")
	} else {
		l.Info("Using local files...")
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
		app.ShutdownModules()
		os.Exit(0)
	}()

	app.RegisterModule(mauth.Module)
	app.RegisterModule(mmain.Module)
	app.RegisterModule(mpref.Module)
	app.RegisterModule(mfiles.Module)
	app.RegisterModule(mmpd.Module)
	app.RegisterModule(mnet.Module)
	app.RegisterModule(mnet.NFSModule)
	app.RegisterModule(mnet.SambaModule)
	app.RegisterModule(mnotepad.Module)
	app.RegisterModule(mstorage.Module)
	app.RegisterModule(msmart.Module)
	app.RegisterModule(msyslogs.Module)
	app.RegisterModule(msyshw.Module)
	app.RegisterModule(msysproc.Module)
	app.RegisterModule(msysusers.Module)
	app.RegisterModule(mutls.Module)
	app.RegisterModule(msystem.Module)
	app.RegisterModule(mmonitor.Module)
	app.RegisterModule(mworker.Module)
	app.RegisterModule(msensors.Module)
	app.InitModules(conf)

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
		l.Info("Listen: %s", conf.HTTPSAddress)
		go func() {
			if err := http.ListenAndServeTLS(conf.HTTPSAddress,
				conf.SslCert, conf.SslKey, nil); err != nil {
				log.Fatalf("Error listening https, %v", err)
			}
		}()
	}

	if conf.HTTPAddress != "" {
		l.Info("Listen: %s", conf.HTTPAddress)
		go func() {
			if err := http.ListenAndServe(conf.HTTPAddress, nil); err != nil {
				log.Fatalf("Error listening http, %v", err)
			}
		}()
	}
	done := make(chan bool)
	<-done
}
