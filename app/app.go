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
				log.Printf("RenderTemplate missing template: %s", fullPath)
				return
			}
			templates = append(templates, fullPath)
		}
		ctemplate = template.Must(template.ParseFiles(templates...))
		templatesCache.items[template_path] = ctemplate
	}
	err := ctemplate.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("RenderTemplate execution failed: %s", err)
	}

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
	Session *sessions.Session
}

func GetSessionStore(w http.ResponseWriter, r *http.Request) *SessionStore {
	session, _ := App.store.Get(r, STORE_SESSION)
	return &SessionStore{session}
}

func (store *SessionStore) Get(key string) interface{} {
	return store.Session.Values[key]
}

func (store *SessionStore) Set(key string, value interface{}) {
	store.Session.Values[key] = value
}

func (store *SessionStore) Clear() {
	store.Session.Values = nil
}

func (store *SessionStore) Save(w http.ResponseWriter, r *http.Request) error {
	err := store.Session.Save(r, w)
	helpers.CheckErr(err, "BasePageContext Save Error")
	return err
}

type BasePageContext struct {
	Title          string
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	*SessionStore
}

func NewBasePageContext(title string, w http.ResponseWriter, r *http.Request) *BasePageContext {
	ctx := &BasePageContext{title, w, r, GetSessionStore(w, r)}
	ctx.GetFlashMessage()
	return ctx
}

func (ctx *BasePageContext) GetFlashMessage() []interface{} {
	if flashes := ctx.Session.Flashes(); len(flashes) > 0 {
		ctx.Session.Save(ctx.Request, ctx.ResponseWriter)
		return flashes
	}
	return nil
}

func (ctx *BasePageContext) AddFlashMessage(msg interface{}) {
	ctx.Session.AddFlash(msg)
	ctx.Session.Save(ctx.Request, ctx.ResponseWriter)
}

func (ctx *BasePageContext) SessionGet(key string) interface{} {
	return ctx.Session.Values[key]
}

func (ctx *BasePageContext) SessionSet(key string, value interface{}) {
	ctx.Session.Values[key] = value
}

func (ctx *BasePageContext) SessionClear() {
	ctx.Session.Values = nil
}
func (ctx *BasePageContext) SessionSave() error {
	err := ctx.Session.Save(ctx.Request, ctx.ResponseWriter)
	helpers.CheckErr(err, "BasePageContext Save Error")
	return err
}
