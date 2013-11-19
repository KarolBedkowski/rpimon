package app

import (
	"../database"
	"../helpers"
	"github.com/coopernurse/gorp"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/keep94/weblogs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// context, sessions
const CONTEXT_APP = "APP"
const STORE_SESSION = "SESSION"
const STORE_FLASH = "flash-session"

type WebApp struct {
	Configuration *AppConfiguration
	Router        *mux.Router
	store         *sessions.CookieStore
	Database      *gorp.DbMap
}

var App *WebApp

func NewWebApp(appConfFile string) *WebApp {

	conf := new(AppConfiguration)
	conf.LoadConfiguration(appConfFile)

	if len(conf.CookieAuthKey) < 32 {
		log.Print("Random CookieAuthKey")
		conf.CookieAuthKey = string(securecookie.GenerateRandomKey(32))
	}
	if len(conf.CookieEncKey) < 32 {
		log.Print("Random CookieEncKey")
		conf.CookieEncKey = string(securecookie.GenerateRandomKey(32))
	}

	app := &WebApp{
		Router:        mux.NewRouter(),
		Configuration: conf,
		store: sessions.NewCookieStore([]byte(conf.CookieAuthKey),
			[]byte(conf.CookieEncKey)),
		Database: database.Init(conf.Database)}

	contextHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer context.Clear(r)
			context.Set(r, CONTEXT_APP, app)
			h.ServeHTTP(w, r)
		})
	}

	http.Handle("/static/", http.StripPrefix("/static",
		http.FileServer(http.Dir(conf.StaticDir))))
	http.Handle("/", weblogs.Handler(contextHandler(csrfHandler(app.Router))))

	App = app
	return app
}

func (app *WebApp) Close() {
	log.Print("Closing...")
	app.Database.Db.Close()
}

type Cache struct {
	mu    sync.Mutex
	items map[string]*template.Template
}

var templatesCache = &Cache{items: make(map[string]*template.Template)}

func (app *WebApp) RenderTemplate(w http.ResponseWriter, name string, data interface{}, filenames ...string) {
	templatesCache.mu.Lock()
	defer templatesCache.mu.Unlock()

	template_path := strings.Join(filenames, "|")

	ctemplate, ok := templatesCache.items[template_path]
	if !ok || App.Configuration.Debug {
		templates := []string{}
		for _, filename := range filenames {
			fullPath := filepath.Join(app.Configuration.TemplatesDir, filename)
			if !fileExists(fullPath) {
				log.Fatalf("RenderTemplate missing template: %s", fullPath)
			}
			templates = append(templates, fullPath)
		}
		ctemplate = template.Must(template.ParseFiles(templates...))
		templatesCache.items[template_path] = ctemplate
	}
	err := ctemplate.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Fatalf("RenderTemplate execution failed: %s", err)
	}

}

func (app *WebApp) GetFlashMessage(w http.ResponseWriter, r *http.Request) []interface{} {
	session, _ := App.store.Get(r, STORE_FLASH)
	if flashes := session.Flashes(); len(flashes) > 0 {
		session.Save(r, w)
		return flashes
	}
	return nil
}

func (app *WebApp) AddFlashMessage(w http.ResponseWriter, r *http.Request, message string) {
	session, _ := app.store.Get(r, STORE_FLASH)
	session.AddFlash(message)
	session.Save(r, w)
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

type SessionStore struct {
	session        *sessions.Session
	responseWriter http.ResponseWriter
	request        *http.Request
}

func NewSessionStore(w http.ResponseWriter, r *http.Request) *SessionStore {
	store, _ := App.store.Get(r, STORE_SESSION)
	return &SessionStore{store, w, r}
}

func (sessStore *SessionStore) Get(key string) interface{} {
	return sessStore.session.Values[key]
}

func (sessStore *SessionStore) Set(key string, value interface{}) {
	sessStore.session.Values[key] = value
}

func (sessStore *SessionStore) Clear() {
	sessStore.session.Values = nil
}
func (sessStore *SessionStore) Save() error {
	err := sessStore.session.Save(sessStore.request, sessStore.responseWriter)
	helpers.CheckErr(err, "SessionStore Save Error")
	return err
}
