package mpd

// MPD current playlist

import (
	"code.google.com/p/gompd/mpd"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strconv"
)

var decoder = schema.NewDecoder()

type playlistPageCtx struct {
	*app.BasePageContext
	CurrentPage   string
	Playlist      []mpd.Attrs
	CurrentSongID string
	CurrentSong   string
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
	if r.Method == "PUT" {
		encoded, _ := json.Marshal(getStatus())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(encoded)
	} else {
		if err != nil {
			session := app.GetSessionStore(w, r)
			session.AddFlash(err.Error())
			session.Save(r, w)
		}
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
	}
}

func playlistActionPageHandler(w http.ResponseWriter, r *http.Request) {
	l.Info("playlistActionPageHandler")
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if ok && action != "" {
		playlistAction(action)
	}
	http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
}

type savePlaylistForm struct {
	Name      string
	CsrfToken string
}

func playlistSavePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	form := &savePlaylistForm{}
	decoder.Decode(form, r.Form)
	if form.Name == "" {
		handleError("Missing name", w, r)
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
		return
	}
	if err := playlistSave(form.Name); err != nil {
		handleError("Saving playlist error: "+err.Error(), w, r)
	} else {
		session := app.GetSessionStore(w, r)
		session.AddFlash("Playlist saved")
		session.Save(r, w)
	}
	http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
}

func handleError(msg string, w http.ResponseWriter, r *http.Request) {
	session := app.GetSessionStore(w, r)
	session.AddFlash(msg)
	session.Save(r, w)
}

type addToPlaylistForm struct {
	Uri       string
	CsrfToken string
}

func addToPlaylistPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	form := &addToPlaylistForm{}
	decoder.Decode(form, r.Form)
	if form.Uri == "" {
		handleError("Missing uri", w, r)
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
		return
	}
	if err := addToPlaylist(form.Uri); err != nil {
		handleError("Adding to playlist error "+err.Error(), w, r)
	} else {
		session := app.GetSessionStore(w, r)
		session.AddFlash("Added to playlist")
		session.Save(r, w)
	}
	http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
}

func sInfoPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	result := map[string]interface{}{"error": nil,
		"aaData":               nil,
		"stat":                 nil,
		"iTotalDisplayRecords": 0,
		"iTotalRecords":        0,
	}
	r.ParseForm()
	var echo = ""
	if echoL, ok := r.Form["sEcho"]; ok {
		echo = echoL[0]
	}
	playlist, err, stat := mpdPlaylistInfo(-1, -1)
	if err == nil {
		for _, item := range playlist {
			if _, ok := item["Artist"]; !ok {
				item["Artist"] = ""
			}
			if _, ok := item["Album"]; !ok {
				item["Album"] = ""
			}
			if _, ok := item["Track"]; !ok {
				item["Track"] = ""
			}
			if title, ok := item["Title"]; !ok || title == "" {
				item["Title"] = item["file"]
			}
		}
		result["stat"] = stat
		result["aaData"] = playlist
		result["sEcho"] = echo
		result["iTotalRecords"] = len(playlist)
		result["iTotalDisplayRecords"] = len(playlist)
	} else {
		result["error"] = err.Error()
	}
	encoded, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}
