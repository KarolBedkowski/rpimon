package mpd

import (
	"code.google.com/p/gompd/mpd"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strconv"
)

var subRouter *mux.Router

// CreateRoutes for /mpd
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler))
	subRouter.HandleFunc("/info", app.VerifyLogged(mainPageHandler)).Name("mpd-index")
	subRouter.HandleFunc("/action/{action}", app.VerifyLogged(actionPageHandler))
	subRouter.HandleFunc("/playlist", app.VerifyLogged(playlistPageHandler)).Name("mpd-playlist")
	subRouter.HandleFunc("/song/{song-id:[0-9]+}/{action}",
		app.VerifyLogged(songActionPageHandler))
	subRouter.HandleFunc("/playlists", app.VerifyLogged(playlistsPageHandler)).Name("mpd-playlists")
	subRouter.HandleFunc("/playlist/{plist}/{action}",
		app.VerifyLogged(playlistsActionPageHandler))
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Status      *mpdStatus
}

var localMenu []app.MenuItem

func createLocalMenu() []app.MenuItem {
	if localMenu == nil {
		localMenu = []app.MenuItem{app.NewMenuItemFromRoute("Info", "mpd-index"),
			app.NewMenuItemFromRoute("Playlist", "mpd-playlist"),
			app.NewMenuItemFromRoute("Playlists", "mpd-playlists")}
	}
	return localMenu
}

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	ctx.CurrentMainMenuPos = "/mpd/"
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentLocalMenuPos = "mpd-index"
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPageCtx(w, r)
	data.Status = getStatus()
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/index.tmpl", "flash.tmpl")
}

func actionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd actionPageHandler: missing action ", vars)
		return
	}
	mpdAction(action)
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
	ctx := &playlistPageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	ctx.CurrentMainMenuPos = "/mpd/"
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
		session.Session.AddFlash(err.Error())
		session.Save(w, r)
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
	ctx := &playlistsPageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	ctx.CurrentMainMenuPos = "/mpd/"
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
		session.Session.AddFlash(err.Error())
		session.Save(w, r)
	}
	http.Redirect(w, r, app.GetNamedURL("mpd-playlists"), http.StatusFound)
}
