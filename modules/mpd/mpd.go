package mpd

import (
	//"code.google.com/p/gompd/mpd"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/turbowookie/gompd/mpd"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/session"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/logging"
	"k.prv/rpimon/model"
	n "k.prv/rpimon/modules/notepad"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Module information
var Module *context.Module

func init() {
	Module = &context.Module{
		Name:        "mpd",
		Title:       "MPD",
		Description: "",
		Init:        initModule,
		GetMenu:     getMenu,
		Shutdown:    shutdown,
		Defaults: map[string]string{
			"host":                     "localhost:6600",
			"log to notes":             "no",
			"delete older than [days]": "0",
			"dump deleted to file":     "",
		},
		Configurable:  true,
		AllPrivilages: []context.Privilege{{"mpd", "manage mpd player"}},
	}
}

// CreateRoutes for /mpd
func initModule(parentRoute *mux.Route) bool {
	conf := Module.GetConfiguration()
	if host, ok := conf["host"]; ok && host != "" {
		initConnector(conf)
	} else {
		l.Warn("MPD missing 'host' configuration parameter")
		return false
	}

	subRouter := parentRoute.Subrouter()
	// Main page
	subRouter.HandleFunc("/", context.SecContext(mainPageHandler, "MPD", "mpd"))
	subRouter.HandleFunc("/main",
		context.SecContext(mainPageHandler, "MPD", "mpd")).Name(
		"mpd-index")
	// Playing control
	subRouter.HandleFunc("/control/{action}",
		app.VerifyPermission(mpdControlHandler, "mpd")).Name(
		"mpd-control")
	// current Playlist
	subRouter.HandleFunc("/playlist",
		context.SecContext(playlistPageHandler, "MPD - Playlist", "mpd")).Name(
		"mpd-playlist")
	subRouter.HandleFunc("/playlist/save",
		app.VerifyPermission(playlistSavePageHandler, "mpd")).Name(
		"mpd-pl-save").Methods("POST")
	subRouter.HandleFunc("/playlist/add",
		app.VerifyPermission(addToPlaylistActionHandler, "mpd")).Name(
		"mpd-pl-add").Methods("POST")
	subRouter.HandleFunc("/playlist/{action}",
		app.VerifyPermission(playlistActionPageHandler, "mpd")).Name(
		"mpd-pl-action")
	subRouter.HandleFunc("/playlist/serv/info",
		app.VerifyPermission(plistContentServHandler, "mpd")).Name(
		"mpd-pl-serv-info")
	subRouter.HandleFunc("/song/{song-id:[0-9]+}/{action}",
		app.VerifyPermission(songActionPageHandler, "mpd")).Name(
		"mpd-song-action")
	// Playlists
	subRouter.HandleFunc("/playlists",
		context.SecContext(playlistsPageHandler, "MPD - Playlists", "mpd")).Name(
		"mpd-playlists")
	subRouter.HandleFunc("/playlists/serv/list",
		app.VerifyPermission(playlistsListService, "mpd")).Name(
		"mpd-playlists-serv-list")
	subRouter.HandleFunc("/playlists/action",
		app.VerifyPermission(playlistsActionPageHandler, "mpd")).Name(
		"mpd-playlists-action")
	// Services
	subRouter.HandleFunc("/service/status",
		app.VerifyPermission(statusServHandler, "mpd")).Name(
		"mpd-service-status")
	subRouter.HandleFunc("/service/song-info",
		app.VerifyPermission(songInfoStubHandler, "mpd")).Name(
		"mpd-service-song-info")
	// Library
	subRouter.HandleFunc("/library",
		context.SecContext(libraryPageHandler, "MPD - Library", "mpd")).Name(
		"mpd-library")
	subRouter.HandleFunc("/library/serv/content",
		app.VerifyPermission(libraryServHandler, "mpd")).Name(
		"mpd-library-content")
	subRouter.HandleFunc("/library/action",
		app.VerifyPermission(libraryActionHandler, "mpd")).Methods(
		"PUT", "POST").Name("mpd-library-action")
	// other
	subRouter.HandleFunc("/log",
		app.VerifyPermission(mpdLogPageHandler, "mpd")).Name(
		"mpd-log")
	// search
	subRouter.HandleFunc("/search",
		context.SecContext(searchPageHandler, "MPD - Search", "mpd")).Name(
		"mpd-search")
	// files
	subRouter.HandleFunc("/file",
		app.VerifyPermission(filePageHandler, "mpd")).Name(
		"mpd-file")
	// history
	subRouter.HandleFunc("/history",
		context.SecContext(historyHandler, "MPD - History", "mpd")).Name("mpd-history")
	subRouter.HandleFunc("/history/file",
		app.VerifyPermission(historyFileHandler, "mpd")).Name("mpd-history-file")

	if val, ok := conf["delete older than [days]"]; ok && val != "0" && val != "" {
		if off, err := strconv.Atoi(val); err == nil {
			maxage := time.Now().Add(time.Duration(off*-24) * time.Hour)
			if filename, ok := conf["dump deleted to file"]; ok && filename != "" {
				model.DumpOldSongsToFile(maxage, filename, true)
			} else {
				model.DeleteOldSongs(maxage)
			}
		}
	}

	return true
}

func getMenu(ctx *context.BasePageContext) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "mpd") {
		return "", nil
	}

	menu = context.NewMenuItem("MPD", "").SetID("mpd").SetIcon("glyphicon glyphicon-music")
	menu.AddChild(
		context.NewMenuItem("Status", app.GetNamedURL("mpd-index")).SetIcon("glyphicon glyphicon-music").SetSortOrder(-2).SetID("mpd-index"),
		context.NewMenuItem("Playlist", app.GetNamedURL("mpd-playlist")).SetIcon("glyphicon glyphicon-list").SetSortOrder(-1).SetID("mpd-playlist"),
		context.NewMenuItem("Library", app.GetNamedURL("mpd-library")).SetIcon("glyphicon glyphicon-folder-open").SetID("mpd-library"),
		context.NewMenuItem("Search", app.GetNamedURL("mpd-search")).SetIcon("glyphicon glyphicon-search").SetID("mpd-search"),
		context.NewMenuItem("Playlists", app.GetNamedURL("mpd-playlists")).SetIcon("glyphicon glyphicon-floppy-open").SetID("mpd-playlists"),
		context.NewMenuItem("Tools", "").SetIcon("glyphicon glyphicon-wrench").SetID("mpd-tools").AddChild(
			context.NewMenuItem("Log", app.GetNamedURL("mpd-log")).SetID("mpd-log"),
			context.NewMenuItem("History", app.GetNamedURL("mpd-history")).SetID("mpd-history"),
		))
	return "", menu
}

func shutdown() {
	closeConnector()
}

var errBadRequest = errors.New("bad request")

type pageCtx struct {
	*context.BasePageContext
	Status *mpdStatus
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &pageCtx{BasePageContext: bctx}
	ctx.SetMenuActive("mpd-index")
	app.RenderTemplateStd(w, ctx, "mpd/index.tmpl")
}

func mpdControlHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd mpdControlHandler: missing action ", vars)
		return
	}
	r.ParseForm()
	err := errBadRequest
	var result = "OK"
	switch action {
	case "volume":
		if vol := r.FormValue("vol"); vol != "" {
			var volInt int
			if volInt, err = strconv.Atoi(vol); err == nil {
				err = setVolume(volInt)
				break
			}
		}
	case "seek":
		if time := r.FormValue("time"); time != "" {
			var timeInt int
			if timeInt, err = strconv.Atoi(time); err == nil {
				err = seekPos(-1, timeInt)
				break
			}
		}
	case "add_to_notes":
		status := getStatus()
		data := make([]string, 0)
		for key, val := range status.Current {
			data = append(data, fmt.Sprintf("%s: %s", key, val))
		}
		data = append(data, "\n-----------------\n\n")
		err = n.AppendToNote("mpd_notes.txt", strings.Join(data, "\n"))
		if err == nil {
			result = "Added to notes"
		}
	case "playlist-clear":
		if err = mpdAction(action); err != nil {
			s := session.GetSessionStore(w, r)
			s.AddFlash(err.Error(), "error")
			session.SaveSession(w, r)
		}
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
	default:
		err = mpdAction(action)
	}

	if err == nil {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.Write([]byte(result))
	} else {
		app.Render400(w, r, "Invalid Request:  "+err.Error())
	}
}

var statusServCache = h.NewSimpleCache(1)

func statusServHandler(w http.ResponseWriter, r *http.Request) {
	data := statusServCache.Get(func() h.Value {
		status := getStatus()
		encoded, _ := json.Marshal(status)
		return encoded
	}).([]byte)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

/*
const fakeResult = `{"Status":{"consume":"0","mixrampdb":"0.000000","mixrampdelay":"nan","nextsong":"1","nextsongid":"1","playlist":"2","playlistlength":"222","random":"0","repeat":"0","single":"0","song":"0","songid":"0","state":"stop","volume":"100","xfade":"0"},"Current":{"Album":"Café Del Mar - Classic I","Artist":"Jules Massenet","Date":"2002","Genre":"Baroque, Modern, Romantic, Classical","Id":"0","Last-Modified":"2013-09-27T06:14:59Z","Pos":"0","Time":"312","Title":"Meditation","Track":"01/12","file":"muzyka/mp3/cafe del mar/compilations/classics/2002, classic/01. jules massenet - meditation.mp3"},"Error":""}`
*/
func songInfoStubHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var ctx struct {
		Error string
		Info  []mpd.Attrs
	}

	if songUri, ok := r.Form["uri"]; ok && songUri[0] != "" {
		uri, _ := url.QueryUnescape(songUri[0])
		result, err := getSongInfo(uri)
		ctx.Info = result
		if err != nil {
			ctx.Error = err.Error()
		}
	}
	app.RenderTemplate(w, ctx, "song-info", "mpd/songinfo.tmpl")
}

func filePageHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	if action == "" {
		app.Render400(w, r, "Invalid Request:  missing action")
		return
	}
	uri := r.FormValue("uri")
	if uri == "" {
		app.Render400(w, r, "Invalid Request:  missing uri")
		return
	}
	uri, _ = url.QueryUnescape(uri)
	err := mpdFileAction(uri, action)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoded, _ := json.Marshal("OK")
		w.Write(encoded)
	} else {
		app.Render400(w, r, "Internal Server Errror: "+err.Error())
	}
}

func historyHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &struct {
		*context.BasePageContext
		History []*model.Song
	}{
		BasePageContext: bctx,
		History:         model.GetSongs(),
	}
	ctx.SetMenuActive("mpd-history")
	app.RenderTemplateStd(w, ctx, "mpd/history.tmpl")
}

func historyFileHandler(w http.ResponseWriter, r *http.Request) {
	songs := model.GetSongs()
	if len(songs) == 0 {
		app.Render400(w, r, "No songs")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"mpd_history.txt\"")
	for _, song := range songs {
		writeNonEmptyString(w, "Date: ", song.Date.String())
		writeNonEmptyString(w, "Track: ", song.Track)
		writeNonEmptyString(w, "Name: ", song.Name)
		writeNonEmptyString(w, "Album: ", song.Album)
		writeNonEmptyString(w, "Artist: ", song.Artist)
		writeNonEmptyString(w, "Title: ", song.Title)
		writeNonEmptyString(w, "File: ", song.File)
		w.Write([]byte("---------------------\n\n"))
	}
}

func writeNonEmptyString(w http.ResponseWriter, prefix, value string) {
	if value != "" {
		w.Write([]byte(prefix + value + "\n"))
	}
}
