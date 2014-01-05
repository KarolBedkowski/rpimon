package mpd

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strconv"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /mpd
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "mpd"))
	subRouter.HandleFunc("/info", app.VerifyPermission(mainPageHandler, "mpd")).Name("mpd-index")
	subRouter.HandleFunc("/action/{action}", app.VerifyPermission(actionPageHandler, "mpd")).Name("mpd-action")
	subRouter.HandleFunc("/log", app.VerifyPermission(mpdLogPageHandler, "mpd")).Name("mpd-log")
	subRouter.HandleFunc("/playlist", app.VerifyPermission(playlistPageHandler, "mpd")).Name("mpd-playlist")
	subRouter.HandleFunc("/song/{song-id:[0-9]+}/{action}",
		app.VerifyPermission(songActionPageHandler, "mpd"))
	subRouter.HandleFunc("/playlists", app.VerifyPermission(playlistsPageHandler, "mpd")).Name("mpd-playlists")
	subRouter.HandleFunc("/playlist/{plist}/{action}",
		app.VerifyPermission(playlistsActionPageHandler, "mpd")).Name("mpd-playlists-action")
	subRouter.HandleFunc("/service/info", app.VerifyPermission(infoHandler, "mpd"))
	subRouter.HandleFunc("/actions", app.VerifyPermission(actionsPageHandler, "mpd")).Name("mpd-actions")
	subRouter.HandleFunc("/library", app.VerifyPermission(libraryPageHandler, "mpd")).Name("mpd-library")
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Status      *mpdStatus
}

var localMenu []*app.MenuItem

func createLocalMenu() []*app.MenuItem {
	if localMenu == nil {
		localMenu = []*app.MenuItem{app.NewMenuItemFromRoute("Info", "mpd-index"),
			app.NewMenuItemFromRoute("Playlist", "mpd-playlist"),
			app.NewMenuItemFromRoute("Library", "mpd-library"),
			app.NewMenuItemFromRoute("Playlists", "mpd-playlists"),
			app.NewMenuItemFromRoute("Log", "mpd-log"),
			app.NewMenuItemFromRoute("Actions", "mpd-actions"),
		}
	}
	return localMenu
}

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r)}
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentLocalMenuPos = "mpd-index"
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPageCtx(w, r)
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/index.tmpl", "flash.tmpl")
}

func actionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd actionPageHandler: missing action ", vars)
		return
	}
	r.ParseForm()
	switch action {
	case "volume":
		{
			if vol := r.Form["vol"][0]; vol != "" {
				volInt, ok := strconv.Atoi(vol)
				if ok == nil {
					setVolume(volInt)
					return
				}
			}
		}
	case "seek":
		{
			if time := r.Form["time"][0]; time != "" {
				timeInt, ok := strconv.Atoi(time)
				if ok == nil {
					seekPos(-1, timeInt)
					return
				}
			}
		}
	default:
		mpdAction(action)
	}
	switch action {
	case "update":
		{
			http.Redirect(w, r, app.GetNamedURL("mpd-index"), http.StatusFound)
		}
	case "playlist-clear":
		{
			http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
		}
	}
}

func mpdLogPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewSimpleDataPageCtx(w, r, "mpd", "mpd", "", createLocalMenu())
	ctx.CurrentLocalMenuPos = "mpd-log"
	ctx.LocalMenu = createLocalMenu()

	if lines, err := h.ReadFromFileLastLines("/var/log/mpd/mpd.log", 25); err != nil {
		ctx.Data = err.Error()
	} else {
		ctx.Data = lines
	}
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}

var infoHandlerCache = h.NewSimpleCache(1)

func infoHandler(w http.ResponseWriter, r *http.Request) {
	data := infoHandlerCache.Get(func() h.Value {
		status := getStatus()
		encoded, _ := json.Marshal(status)
		return encoded
	}).([]byte)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

const fakeResult = `{"Status":{"consume":"0","mixrampdb":"0.000000","mixrampdelay":"nan","nextsong":"1","nextsongid":"1","playlist":"2","playlistlength":"222","random":"0","repeat":"0","single":"0","song":"0","songid":"0","state":"stop","volume":"100","xfade":"0"},"Current":{"Album":"Café Del Mar - Classic I","Artist":"Jules Massenet","Date":"2002","Genre":"Baroque, Modern, Romantic, Classical","Id":"0","Last-Modified":"2013-09-27T06:14:59Z","Pos":"0","Time":"312","Title":"Meditation","Track":"01/12","file":"muzyka/mp3/cafe del mar/compilations/classics/2002, classic/01. jules massenet - meditation.mp3"},"Error":""}`

func actionsPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newPageCtx(w, r)
	ctx.CurrentLocalMenuPos = "mpd-actions"
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "mpd/actions.tmpl", "flash.tmpl")
}

type libraryPageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Path        string
	Files       []string
	Folders     []string
	Error       string
}

func libraryPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &libraryPageCtx{BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r)}
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentLocalMenuPos = "mpd-library"
	ctx.Path = ""
	r.ParseForm()
	if path, ok := r.Form["p"]; ok {
		ctx.Path = strings.TrimLeft(path[0], "/")
	}
	if action, ok := r.Form["a"]; ok {
		switch action[0] {
		case "add":
			{
				if uriL, ok := r.Form["u"]; ok {
					uri := strings.TrimLeft(uriL[0], "/")
					err := addFileToPlaylist(uri, false)
					if err == nil {
						ctx.AddFlashMessage("Added " + uri)
						ctx.Save()
						http.Redirect(w, r, app.GetNamedURL("mpd-library")+
							h.BuildQuery("p", ctx.Path), http.StatusFound)
						return
					} else {
						ctx.Error = err.Error()
					}
				}
			}
		case "replace":
			{
				if uriL, ok := r.Form["u"]; ok {
					uri := strings.TrimLeft(uriL[0], "/")
					err := addFileToPlaylist(uri, true)
					if err == nil {
						ctx.AddFlashMessage("Playlist cleared and added " + uri)
						ctx.Save()
						http.Redirect(w, r, app.GetNamedURL("mpd-library")+
							h.BuildQuery("p", ctx.Path), http.StatusFound)
						return
					} else {
						ctx.Error = err.Error()
					}
				}
			}
		case "up":
			{
				if ctx.Path != "" {
					idx := strings.LastIndex(ctx.Path, "/")
					if idx > 0 {
						ctx.Path = ctx.Path[:idx]
					} else {
						ctx.Path = ""
					}
				}
			}
		}
	}
	ctx.Folders, ctx.Files, _ = getFiles(ctx.Path)
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "mpd/library.tmpl", "flash.tmpl")
}
