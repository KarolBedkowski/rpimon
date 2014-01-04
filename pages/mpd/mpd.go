package mpd

import (
	"code.google.com/p/gompd/mpd"
	"encoding/json"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strconv"
)

var subRouter *mux.Router

// CreateRoutes for /mpd
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "mpd"))
	subRouter.HandleFunc("/info", app.VerifyPermission(mainPageHandler, "mpd")).Name("mpd-index")
	subRouter.HandleFunc("/action/{action}", app.VerifyPermission(actionPageHandler, "mpd"))
	subRouter.HandleFunc("/log", app.VerifyPermission(mpdLogPageHandler, "mpd")).Name("mpd-log")
	subRouter.HandleFunc("/playlist", app.VerifyPermission(playlistPageHandler, "mpd")).Name("mpd-playlist")
	subRouter.HandleFunc("/song/{song-id:[0-9]+}/{action}",
		app.VerifyPermission(songActionPageHandler, "mpd"))
	subRouter.HandleFunc("/playlists", app.VerifyPermission(playlistsPageHandler, "mpd")).Name("mpd-playlists")
	subRouter.HandleFunc("/playlist/{plist}/{action}",
		app.VerifyPermission(playlistsActionPageHandler, "mpd"))
	subRouter.HandleFunc("/service/info", app.VerifyPermission(infoHandler, "mpd"))
	subRouter.HandleFunc("/actions", app.VerifyPermission(actionsPageHangler, "mpd")).Name("mpd-actions")
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
	http.Redirect(w, r, app.GetNamedURL("mpd-index"), http.StatusFound)
}

type playlistPageCtx struct {
	*app.BasePageContext
	CurrentPage   string
	Playlist      []mpd.Attrs
	CurrentSongID string
	Error         error
}

func newPlaylistPageCtx(w http.ResponseWriter, r *http.Request) *playlistPageCtx {
	ctx := &playlistPageCtx{BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r)}
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentLocalMenuPos = "mpd-playlist"
	return ctx
}

func playlistPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPlaylistPageCtx(w, r)
	playlist, err, current := mpdPlaylistInfo()
	data.Playlist = playlist
	data.Error = err
	data.CurrentSongID = current
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/playlist.tmpl", "flash.tmpl")
}

func songActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd songActionPageHandler: missing action ", vars)
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
		return
	}
	songIDStr, ok := vars["song-id"]
	if !ok || songIDStr == "" {
		l.Warn("page.mpd songActionPageHandler: missing songID ", vars)
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
		return
	}
	songID, err := strconv.Atoi(songIDStr)
	if err != nil || songID < 0 {
		l.Warn("page.mpd songActionPageHandler: wrong songID ", vars)
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
		return
	}
	err = mpdSongAction(songID, action)
	if err != nil {
		session := app.GetSessionStore(w, r)
		session.AddFlash(err.Error())
		session.Save(r, w)
	}
	http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
}

type playlistsPageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Playlists   []mpd.Attrs
	Error       error
}

func newPlaylistsPageCtx(w http.ResponseWriter, r *http.Request) *playlistsPageCtx {
	ctx := &playlistsPageCtx{BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r)}
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentLocalMenuPos = "mpd-playlists"
	return ctx
}
func playlistsPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPlaylistsPageCtx(w, r)
	playlists, err := mpdGetPlaylists()
	data.Playlists = playlists
	data.Error = err
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/playlists.tmpl", "flash.tmpl")
}

func playlistsActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd playlistsActionPageHandler: missing action ", vars)
		http.Redirect(w, r, app.GetNamedURL("mpd-playlists"), http.StatusFound)
		return
	}
	playlist, ok := vars["plist"]
	if !ok || playlist == "" {
		l.Warn("page.mpd playlistsActionPageHandler: missing songID ", vars)
		http.Redirect(w, r, app.GetNamedURL("mpd-playlists"), http.StatusFound)
		return
	}
	err := mpdPlaylistAction(playlist, action)
	if err != nil {
		session := app.GetSessionStore(w, r)
		session.AddFlash(err.Error())
		session.Save(r, w)
	}
	http.Redirect(w, r, app.GetNamedURL("mpd-playlists"), http.StatusFound)
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

func actionsPageHangler(w http.ResponseWriter, r *http.Request) {
	ctx := newPageCtx(w, r)
	ctx.CurrentLocalMenuPos = "mpd-actions"
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "mpd/actions.tmpl", "flash.tmpl")
}
