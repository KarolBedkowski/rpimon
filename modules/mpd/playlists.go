package mpd

// MPD Playlists

import (
	//"code.google.com/p/gompd/mpd"
	"encoding/json"
	"github.com/turbowookie/gompd/mpd"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/errors"
	h "k.prv/rpimon/helpers"
	//l "k.prv/rpimon/helpers/logging"
	"net/http"
)

type playlistsPageCtx struct {
	*context.BasePageContext
	CurrentPage string
	Playlists   []mpd.Attrs
	Error       string
}

func playlistsPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &playlistsPageCtx{BasePageContext: bctx}
	ctx.SetMenuActive("mpd-playlists")
	app.RenderTemplateStd(w, ctx, "mpd/playlists.tmpl")
}

func playlistsActionPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
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
		errors.Render500(w, r, "Playlist action error: "+err.Error())
	}
}

func playlistsListService(w http.ResponseWriter, r *http.Request) {
	result := make(map[string]interface{})
	if playlists, err := mpdGetPlaylists(); err != nil {
		result["error"] = err.Error()
	} else {
		result["items"] = playlists
		result["error"] = ""
	}
	encoded, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}
