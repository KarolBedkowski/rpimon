package mpd

// MPD current playlist

import (
	//"code.google.com/p/gompd/mpd"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/turbowookie/gompd/mpd"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/session"
	"net/http"
	"strconv"
	"strings"
)

var decoder = schema.NewDecoder()

type playlistPageCtx struct {
	*context.BasePageContext
	Playlist      []mpd.Attrs
	CurrentSongID string
	CurrentSong   string
	Error         error
}

func playlistPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &playlistPageCtx{BasePageContext: bctx}
	ctx.SetMenuActive("mpd-playlist")
	app.RenderTemplateStd(w, ctx, "mpd/playlist.tmpl")
}

func songActionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		app.Render400(w, r, "Invalid Request: missing action")
		return
	}
	var songID = -2
	if songIDStr, ok := vars["song-id"]; ok && songIDStr != "" {
		songID, _ = strconv.Atoi(songIDStr)
	}
	if songID == -2 {
		app.Render400(w, r, "Invalid Request: missing or invalid songID")
		return
	}
	err := mpdSongAction(songID, action)
	if r.Method == "PUT" {
		encoded, _ := json.Marshal(getStatus())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(encoded)
	} else {
		if err != nil {
			s := session.GetSessionStore(w, r)
			s.AddFlash(err.Error(), "error")
			s.Save(r, w)
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
		app.Render400(w, r, "Invalid Request: missing name")
		return
	}
	if err := playlistSave(form.Name); err != nil {
		app.Render500(w, r, "Saving playlist error: "+err.Error())
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
		app.Render400(w, r, "Invalid Request: missing URI")
		return
	}
	if err := addToPlaylist(form.Uri); err != nil {
		app.Render500(w, r, "Adding to playlist error: "+err.Error())
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
