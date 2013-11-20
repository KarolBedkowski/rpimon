package main

import (
	"./app"
	"./users"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	configFilename := flag.String("conf", "./config.json", "Configuration filename")
	debug := flag.Bool("debug", false, "Run in debug mode")
	httpAddr := flag.String("addr", ":8000", "HTTP server address")
	flag.Parse()

	app.Init(*configFilename, *debug)
	defer app.Close()

	app.Router.HandleFunc("/", handleHome)
	app.Router.HandleFunc("/auth/login", users.LoginHandler).Name("auth-login")
	app.Router.HandleFunc("/auth/logoff", users.LogoffHandler).Name("auth-logoff")
	users.CreateRoutes(app.Router.PathPrefix("/users"))

	log.Printf("Listen: %s", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatalf("Error listening, %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewBasePageContext("Home", w, r)
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "index.tmpl",
		"flash.tmpl")
}
