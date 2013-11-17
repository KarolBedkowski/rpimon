package main

import (
	"./app"
	"./users"
	"flag"
	"log"
	"net/http"
)

func main() {
	wapp := app.Init(".")
	defer wapp.Close()

	app.App.Router.HandleFunc("/", handleHome)
	app.App.Router.HandleFunc("/auth/login", users.LoginHandler).Name("auth-login")
	users.CreateRoutes(app.App.Router.PathPrefix("/users"))

	httpAddr := flag.String("addr", ":8000", "HTTP server address")
	log.Printf("Listen: %s", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatalf("Error listening, %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	app.App.RenderTemplate(w, "base", nil, "base.tmpl", "index.tmpl")
}
