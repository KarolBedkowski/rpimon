package main

import (
	"./app"
	"./users"
	"flag"
	"log"
	"net/http"
)

func main() {
	configFilename := flag.String("conf", "./config.json", "Configuration filename")
	debug := flag.Bool("debug", false, "Run in debug mode")
	httpAddr := flag.String("addr", ":8000", "HTTP server address")
	flag.Parse()

	wapp := app.NewWebApp(*configFilename)
	defer wapp.Close()

	wapp.Configuration.Debug = *debug
	log.Print("Debug=", wapp.Configuration.Debug)

	wapp.Router.HandleFunc("/", handleHome)
	wapp.Router.HandleFunc("/auth/login", users.LoginHandler).Name("auth-login")
	wapp.Router.HandleFunc("/auth/logoff", users.LogoffHandler).Name("auth-logoff")
	users.CreateRoutes(wapp.Router.PathPrefix("/users"))

	log.Printf("Listen: %s", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatalf("Error listening, %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	app.App.RenderTemplate(w, "base", nil, "base.tmpl", "index.tmpl")
}
