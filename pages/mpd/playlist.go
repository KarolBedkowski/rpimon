package mpd

// MPD current playlist

import (
	//"code.google.com/p/gompd/mpd"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/turbowookie/gompd/mpd"
	"k.prv/rpimon/app"
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

func playlistPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &playlistPageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	app.AttachSubmenu(ctx.BasePageContext, "mpd", buildLocalMenu())
	ctx.SetMenuActive("mpd-playlist")
	app.RenderTemplateStd(w, ctx, "mpd/playlist.tmpl")
}

func songActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd.songActionPageHandler: missing action ", vars)
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}
	var songID = -2
	if songIDStr, ok := vars["song-id"]; ok && songIDStr != "" {
		songID, _ = strconv.Atoi(songIDStr)
	}
	if songID == -2 {
		l.Warn("page.mpd.songActionPageHandler: missing or invalid songID ", vars)
		http.Error(w, "missing or invalid songID", http.StatusBadRequest)
		return
	}
	err := mpdSongAction(songID, action)
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
	vars := mux.Vars(r)
	if action, ok := vars["action"]; ok && action != "" {
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

type addToPlaylistForm struct {
	Uri       string
	CsrfToken string
}

func addToPlaylistActionHandler(w http.ResponseWriter, r *http.Request) {
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

func getPlaylistStat() (playlist [][]string, stat mpd.Attrs, err error) {
	lplaylist, err, stat := mpdPlaylistInfo()
	for _, item := range lplaylist {
		if title, ok := item["Title"]; !ok || title == "" {
			item["Title"] = item["file"]
		}
		playlist = append(playlist, []string{item["Album"],
			item["Artist"], item["Track"], item["Title"],
			item["Id"], item["file"],
		})
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

type plistContentParams struct {
	Start  int    `schema:"iDisplayStart"`
	End    int    `schema:"iDisplayLength"`
	Echo   string `schema:"sEcho"`
	Search string `schema:"sSearch"`
}

func plistContentServHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := &plistContentParams{}
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

	if playlist, stat, err := getPlaylistStat(); err == nil {
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
