package mpd

// MPD current playlist

import (
	"code.google.com/p/gompd/mpd"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"strconv"
	"strings"
)

var decoder = schema.NewDecoder()

type playlistPageCtx struct {
	*app.BasePageContext
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
			session.AddFlash(err.Error(), "error")
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
		http.Error(w, "Missing name", http.StatusBadRequest)
		return
	}
	if err := playlistSave(form.Name); err != nil {
		http.Error(w, "Saving playlist error: "+err.Error(), http.StatusInternalServerError)
	} else {
		w.Write([]byte("Playlist saved"))
	}
}

func handleError(msg string, w http.ResponseWriter, r *http.Request) {
	session := app.GetSessionStore(w, r)
	session.AddFlash(msg, "error")
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
		http.Error(w, "Missing URI", http.StatusBadRequest)
		return
	}
	if err := addToPlaylist(form.Uri); err != nil {
		http.Error(w, "Adding to playlist error: "+err.Error(), http.StatusInternalServerError)
	} else {
		w.Write([]byte("URI added"))
	}
}

var mpdPlaylistCache = h.NewSimpleCache(10)
var mpdStatusCache = h.NewSimpleCache(10)

func getPlaylistStat() (playlist [][]string, stat mpd.Attrs, err error) {
	if cachedPlaylist, ok := mpdPlaylistCache.GetValue(); ok {
		playlist = cachedPlaylist.([][]string)
	}
	if cachedStat, ok := mpdStatusCache.GetValue(); ok {
		stat = cachedStat.(mpd.Attrs)
	}
	if len(playlist) == 0 || len(stat) == 0 {
		var lplaylist []mpd.Attrs
		lplaylist, err, stat = mpdPlaylistInfo(-1, -1)
		for _, item := range lplaylist {
			l.Print(item)
			if title, ok := item["Title"]; !ok || title == "" {
				item["Title"] = item["file"]
			}
			playlist = append(playlist, []string{item["Album"],
				item["Artist"], item["Track"], item["Title"],
				item["Id"], item["file"],
			})
		}
	}
	return
}

func filterPlaylist(playlist [][]string, filter string) (filtered [][]string) {
	filtered = make([][]string, 0)
	filter = strings.ToLower(filter)
	for _, item := range playlist {
		for _, value := range item {
			if strings.Contains(strings.ToLower(value), filter) {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return
}

type sInfoParams struct {
	Start  int    `schema:"iDisplayStart"`
	End    int    `schema:"iDisplayLength"`
	Echo   string `schema:"sEcho"`
	Search string `schema:"sSearch"`
}

func sInfoPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := &sInfoParams{}
	decoder.Decode(params, r.Form)
	start := params.Start
	if params.End > 0 {
		params.End += start
	}
	end := params.End
	result := map[string]interface{}{"error": nil,
		"aaData":               nil,
		"stat":                 nil,
		"iTotalDisplayRecords": 0,
		"iTotalRecords":        0,
		"sEcho":                params.Echo,
	}

	playlist, stat, err := getPlaylistStat()

	if err == nil {
		if params.Search != "" {
			playlist = filterPlaylist(playlist, params.Search)
			result["iTotalDisplayRecords"] = len(playlist)
		} else {
			result["iTotalDisplayRecords"] = stat["playlistlength"]
		}
		if len(playlist) > 0 {
			if end > 0 && end < len(playlist) {
				playlist = playlist[start:end]
			} else {
				playlist = playlist[start:]
			}
		}
		result["iTotalRecords"] = stat["playlistlength"]
		result["stat"] = stat
		result["aaData"] = playlist
	} else {
		result["error"] = err.Error()
	}

	encoded, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}
