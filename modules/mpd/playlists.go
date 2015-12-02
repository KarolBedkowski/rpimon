package mpd

// MPD Playlists

import (
	//"code.google.com/p/gompd/mpd"
	"encoding/json"
	"github.com/fhs/gompd/mpd"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/logging"
	"net/http"
	"net/url"
	"strings"
)

type playlistsPageCtx struct {
	*app.BaseCtx
	CurrentPage string
	Playlists   []mpd.Attrs
	Error       string
}

func playlistsPageHandler(w http.ResponseWriter, r *http.Request, bctx *app.BaseCtx) {
	ctx := &playlistsPageCtx{BaseCtx: bctx}
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
		app.Render500(w, r, "Playlist action error: "+err.Error())
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

func playlistsContentPage(w http.ResponseWriter, r *http.Request, bctx *app.BaseCtx) {
	vars := mux.Vars(r)
	playlist, ok := vars["name"]
	if !ok || playlist == "" {
		l.Warn("page.mpd playlistsContentPage: missing name ", vars)
		return
	}
	content, err := mpdGetPlaylistContent(playlist)
	if err != nil {
		l.Warn("page.mpd playlistsContentPage: get content error: %s ", err.Error())
	}
	ctx := struct {
		*app.BaseCtx
		Content []mpd.Attrs
		Name    string
	}{
		BaseCtx: bctx,
		Content: content,
		Name:    playlist,
	}
	ctx.SetMenuActive("mpd-playlists")
	app.RenderTemplateStd(w, ctx, "mpd/playlist_content.tmpl")
}

func playlistsSongActionHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	action := r.FormValue("a")
	switch action {
	case "add", "replace":
		if uriL, ok := r.Form["u"]; ok {
			uri := strings.TrimLeft(uriL[0], "/")
			uri, _ = url.QueryUnescape(uri)
			err := addFileToPlaylist(uri, action == "replace")
			if err == nil {
				w.Write([]byte("Added to playlist"))
			} else {
				l.Error("mpd.playlistsSongActionHandler error: %s", err.Error())
				app.Render400(w, r)
			}
			return
		}
	}
	app.Render400(w, r)
}
