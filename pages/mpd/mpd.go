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

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("mpd-index")
	subRouter.HandleFunc("/info", app.VerifyLogged(mainPageHandler)).Name("mpd-index")
	subRouter.HandleFunc("/action/{action}", app.VerifyLogged(actionPageHandler))
	subRouter.HandleFunc("/playlist", app.VerifyLogged(playlistPageHandler))
	subRouter.HandleFunc("/song/{song-id:[0-9]+}/{action}",
		app.VerifyLogged(songActionPageHandler))
	subRouter.HandleFunc("/playlists", app.VerifyLogged(playlistsPageHandler))
	subRouter.HandleFunc("/playlist/{plist}/{action}",
		app.VerifyLogged(playlistActionPageHandler))
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Status      *mpdStatus
}

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	ctx.CurrentMainMenuPos = "/mpd/"
	ctx.LocalMenu = []app.MenuItem{app.NewMenuItem("Info", "info"),
		app.NewMenuItem("Playlist", "playlist"),
		app.NewMenuItem("Playlists", "playlists")}
	ctx.CurrentLocalMenuPos = "info"
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
		mainPageHandler(w, r)
		return
	}
	data := newPageCtx(w, r)
	data.Status = mpdAction(action)
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/index.tmpl", "flash.tmpl")
}

type playlistPageCtx struct {
	*app.BasePageContext
	CurrentPage   string
	Playlist      []mpd.Attrs
	CurrentSongId string
	Error         error
}

func newPlaylistPageCtx(w http.ResponseWriter, r *http.Request) *playlistPageCtx {
	ctx := &playlistPageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	ctx.CurrentMainMenuPos = "/mpd/"
	ctx.LocalMenu = []app.MenuItem{app.NewMenuItem("Info", "info"),
		app.NewMenuItem("Playlist", "playlist"),
		app.NewMenuItem("Playlists", "playlists")}
	ctx.CurrentLocalMenuPos = "playlist"
	return ctx
}

func playlistPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPlaylistPageCtx(w, r)
	playlist, err, current := mpdPlaylistInfo()
	data.Playlist = playlist
	data.Error = err
	data.CurrentSongId = current
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/playlist.tmpl", "flash.tmpl")
}

func songActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd songActionPageHandler: missing action ", vars)
		playlistPageHandler(w, r)
		return
	}
	songIdStr, ok := vars["song-id"]
	if !ok || songIdStr == "" {
		l.Warn("page.mpd songActionPageHandler: missing songid ", vars)
		playlistPageHandler(w, r)
		return
	}
	songId, err := strconv.Atoi(songIdStr)
	if err != nil || songId < 0 {
		l.Warn("page.mpd songActionPageHandler: wrong songid ", vars)
		playlistPageHandler(w, r)
		return
	}
	err = mpdSongAction(songId, action)
	if err != nil {
		session := app.GetSessionStore(w, r)
		session.Session.AddFlash(err.Error())
		session.Save(w, r)
	}
	http.Redirect(w, r, "/mpd/playlist", http.StatusFound)
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
	ctx.LocalMenu = []app.MenuItem{app.NewMenuItem("Info", "info"),
		app.NewMenuItem("Playlist", "playlist"),
		app.NewMenuItem("Playlists", "playlists")}
	ctx.CurrentLocalMenuPos = "playlists"
	return ctx
}
func playlistsPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPlaylistsPageCtx(w, r)
	playlists, err := mpdGetPlaylists()
	data.Playlists = playlists
	data.Error = err
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/playlists.tmpl", "flash.tmpl")
}

func playlistActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd playlistActionPageHandler: missing action ", vars)
		playlistPageHandler(w, r)
		return
	}
	playlist, ok := vars["plist"]
	if !ok || playlist == "" {
		l.Warn("page.mpd playlistActionPageHandler: missing songid ", vars)
		playlistPageHandler(w, r)
		return
	}
	err := mpdPlaylistAction(playlist, action)
	if err != nil {
		session := app.GetSessionStore(w, r)
		session.Session.AddFlash(err.Error())
		session.Save(w, r)
	}
	http.Redirect(w, r, "/mpd/playlists", http.StatusFound)
}
