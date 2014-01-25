package mpd

// MPD Playlists

import (
	"code.google.com/p/gompd/mpd"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	//	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

type playlistsPageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Playlists   []mpd.Attrs
	Error       string
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
	if err != nil {
		data.Error = err.Error()
	}
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/playlists.tmpl", "flash.tmpl")
}

func playlistsActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}
	playlist, ok := vars["plist"]
	if !ok || playlist == "" {
		http.Error(w, "missing songid", http.StatusBadRequest)
		return
	}
	status, err := mpdPlaylistsAction(playlist, action)
	if err == nil {
		w.Write([]byte(status))
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
