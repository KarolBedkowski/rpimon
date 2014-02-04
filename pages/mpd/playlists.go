package mpd

// MPD Playlists

import (
	"code.google.com/p/gompd/mpd"
	"encoding/json"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	//l "k.prv/rpimon/helpers/logging"
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
	ctx.LocalMenu = localMenu
	ctx.CurrentLocalMenuPos = "mpd-playlists"
	return ctx
}

func playlistsPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPlaylistsPageCtx(w, r)
	app.RenderTemplateStd(w, data, "mpd/playlists.tmpl")
}

func playlistsActionPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//l.Debug("Form: %#v", r.Form)
	action, ok := h.GetParam(w, r, "a")
	if !ok {
		return
	}
	playlist, ok := h.GetParam(w, r, "p")
	if !ok {
		return
	}
	status, err := mpdPlaylistsAction(playlist, action)
	if err == nil {
		w.Write([]byte(status))
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func playlistsListService(w http.ResponseWriter, r *http.Request) {
	result := map[string]interface{}{
		"error": "",
	}
	if playlists, err := mpdGetPlaylists(); err != nil {
		result["error"] = err.Error()
	} else {
		result["items"] = playlists
	}
	encoded, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}
