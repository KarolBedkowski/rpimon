package mpd

// MPD Playlists

import (
	"code.google.com/p/gompd/mpd"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

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
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/playlists.tmpl", "flash.tmpl", "pager.tmpl")
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
	err := mpdPlaylistsAction(playlist, action)
	if err != nil {
		session := app.GetSessionStore(w, r)
		session.AddFlash(err.Error())
		session.Save(r, w)
	}
	if r.Method == "GET" {
		http.Redirect(w, r, app.GetNamedURL("mpd-playlists"), http.StatusFound)
	} else {
		w.Write([]byte("OK"))
	}
}
