package mpd

// MPD current playlist

import (
	"code.google.com/p/gompd/mpd"
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
