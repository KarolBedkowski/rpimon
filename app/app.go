package app

import (
	"../database"
	"crypto/rand"
	"encoding/base64"
	"github.com/coopernurse/gorp"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/keep94/weblogs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

const COOKIE_SECRET = "cookie-secret"
const APP_ID = "my_app"
const APP_SECRET = "my_app-secret"
const BASE_TEMPLATE = "base.tmpl"

type WebApp struct {
	Router      *mux.Router
	StaticDir   string
	TemplateDir string
	CookieStore *sessions.CookieStore
	Database    *gorp.DbMap
}

var App *WebApp

func Init(appRoot string) *WebApp {
	staticDir := filepath.Join(appRoot, "static")
	if staticDir == "" || !fileExists(staticDir) {
		log.Fatal("Missing static dir")
	}
	templateDir := filepath.Join(appRoot, "templates")
	if templateDir == "" || !fileExists(templateDir) {
		log.Fatal("Missing templates dir")
	}

	app := &WebApp{
		Router:      mux.NewRouter(),
		StaticDir:   staticDir,
		TemplateDir: templateDir,
		CookieStore: sessions.NewCookieStore([]byte(COOKIE_SECRET)),
		Database:    database.Init()}

	http.Handle("/static/", http.StripPrefix("/static",
		http.FileServer(http.Dir(staticDir))))
	http.Handle("/", context.ClearHandler(weblogs.Handler(app.Router)))

	App = app
	return app
}

func (app *WebApp) Close() {
	log.Print("Closing...")
	app.Database.Db.Close()
}

const CSRF_TOKEN_KEY = "csrf_token"

func (app *WebApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the session and write a DummyKey
	session, _ := App.CookieStore.Get(r, "SESSION")
	csrf_token, ok := session.Values[CSRF_TOKEN_KEY]
	if !ok {
		csrf_token = generateCsrfToken()
		session.Values[CSRF_TOKEN_KEY] = csrf_token
		App.CookieStore.Save(r, w, session)
	}

	context.Set(r, CSRF_TOKEN_KEY, csrf_token)
	if r.Method == "POST" && r.FormValue("csrf_token") != csrf_token {
		http.Error(w, "Fobidden", http.StatusForbidden)
	} else {
		app.Router.ServeHTTP(w, r)
	}
}

func (app *WebApp) RenderTemplate(w http.ResponseWriter, name string, data interface{}, filenames ...string) {
	templates := []string{}
	for _, filename := range filenames {
		fullPath := filepath.Join(app.TemplateDir, filename)
		if !fileExists(fullPath) {
			log.Fatalf("RenderTemplate missing template: %s", fullPath)
		}
		templates = append(templates, fullPath)
	}
	paresedTemplates, err := template.ParseFiles(templates...)
	if err != nil {
		log.Fatalf("RenderTemplate parse failed: %s", err)
	}
	err = paresedTemplates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Fatalf("RenderTemplate execution failed: %s", err)
	}

}

func logRequests(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func generateCsrfToken() string {
	token := make([]byte, 80)
	rand.Read(token)
	return base64.StdEncoding.EncodeToString(token)
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			log.Fatal(name, " does not exist.")
		}
		return false
	}
	return true
}

const FLASH_SESSION_NAME = "flash-session"

func GetFlashMessage(w http.ResponseWriter, r *http.Request) []interface{} {
	store := App.CookieStore
	session, _ := store.Get(r, FLASH_SESSION_NAME)
	if flashes := session.Flashes(); len(flashes) > 0 {
		session.Save(r, w)
		return flashes
	}
	return nil
}

func AddFlashMessage(w http.ResponseWriter, r *http.Request, message string) {
	store := App.CookieStore
	session, _ := store.Get(r, FLASH_SESSION_NAME)
	session.AddFlash(message)
	session.Save(r, w)
}
