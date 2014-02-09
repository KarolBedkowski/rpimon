package mpd

import (
	//"code.google.com/p/gompd/mpd"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/turbowookie/gompd/mpd"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CreateRoutes for /mpd
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	// Main page
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "mpd"))
	subRouter.HandleFunc("/main",
		app.VerifyPermission(mainPageHandler, "mpd")).Name(
		"mpd-index")
	// Playing control
	subRouter.HandleFunc("/control/{action}",
		app.VerifyPermission(controlHandler, "mpd")).Name(
		"mpd-control")
	// current Playlist
	subRouter.HandleFunc("/playlist",
		app.VerifyPermission(playlistPageHandler, "mpd")).Name(
		"mpd-playlist")
	subRouter.HandleFunc("/playlist/save",
		app.VerifyPermission(playlistSavePageHandler, "mpd")).Name(
		"mpd-pl-save").Methods("POST")
	subRouter.HandleFunc("/playlist/add",
		app.VerifyPermission(addToPlaylistPageHandler, "mpd")).Name(
		"mpd-pl-add").Methods("POST")
	subRouter.HandleFunc("/playlist/{action}",
		app.VerifyPermission(playlistActionPageHandler, "mpd")).Name(
		"mpd-pl-action")
	subRouter.HandleFunc("/playlist/serv/info",
		app.VerifyPermission(sInfoPlaylistHandler, "mpd")).Name(
		"mpd-pl-serv-info")
	subRouter.HandleFunc("/song/{song-id:[0-9]+}/{action}",
		app.VerifyPermission(songActionPageHandler, "mpd")).Name(
		"mpd-song-action")
	// Playlists
	subRouter.HandleFunc("/playlists",
		app.VerifyPermission(playlistsPageHandler, "mpd")).Name(
		"mpd-playlists")
	subRouter.HandleFunc("/playlists/serv/list",
		app.VerifyPermission(playlistsListService, "mpd")).Name(
		"mpd-playlists-serv-list")
	subRouter.HandleFunc("/playlists/action",
		app.VerifyPermission(playlistsActionPageHandler, "mpd")).Name(
		"mpd-playlists-action")
	// Services
	subRouter.HandleFunc("/service/info",
		app.VerifyPermission(infoHandler, "mpd")).Name(
		"mpd-service-info")
	subRouter.HandleFunc("/service/song-info",
		app.VerifyPermission(songInfoHandler, "mpd")).Name(
		"mpd-service-song-info")
	// Library
	subRouter.HandleFunc("/library",
		app.VerifyPermission(libraryPageHandler, "mpd")).Name(
		"mpd-library")
	subRouter.HandleFunc("/library/content",
		app.VerifyPermission(libraryContentService, "mpd")).Name(
		"mpd-library-content")
	subRouter.HandleFunc("/library/action",
		app.VerifyPermission(libraryActionHandler, "mpd")).Methods(
		"PUT", "POST").Name("mpd-library-action")
	// other
	subRouter.HandleFunc("/log",
		app.VerifyPermission(mpdLogPageHandler, "mpd")).Name(
		"mpd-log")
	subRouter.HandleFunc("/notes",
		app.VerifyPermission(notesPageHandler, "mpd")).Name(
		"mpd-notes")
	// search
	subRouter.HandleFunc("/search",
		app.VerifyPermission(searchPageHandler, "mpd")).Name(
		"mpd-search")
	// files
	subRouter.HandleFunc("/file",
		app.VerifyPermission(filePageHandler, "mpd")).Name(
		"mpd-file")
	localMenu = []*app.MenuItem{app.NewMenuItemFromRoute("Status", "mpd-index").SetIcon("glyphicon glyphicon-music"),
		app.NewMenuItemFromRoute("Playlist", "mpd-playlist").SetIcon("glyphicon glyphicon-list"),
		app.NewMenuItemFromRoute("Library", "mpd-library").SetIcon("glyphicon glyphicon-folder-open"),
		app.NewMenuItemFromRoute("Search", "mpd-search").SetIcon("glyphicon glyphicon-search"),
		app.NewMenuItemFromRoute("Playlists", "mpd-playlists").SetIcon("glyphicon glyphicon-floppy-open"),
		app.NewMenuItemFromRoute("Tools", "mpd-tools").SetIcon("glyphicon glyphicon-wrench").AddChild(
			app.NewMenuItemFromRoute("Log", "mpd-log")).AddChild(app.NewMenuItemFromRoute("Notes", "mpd-notes")),
	}
}

type pageCtx struct {
	*app.BasePageContext
	Status *mpdStatus
}

var localMenu []*app.MenuItem

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r)}
	ctx.LocalMenu = localMenu
	ctx.SetMenuActive("mpd-index")
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPageCtx(w, r)
	app.RenderTemplateStd(w, data, "mpd/index.tmpl")
}

func controlHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd controlHandler: missing action ", vars)
		return
	}
	r.ParseForm()
	err := errors.New("invalid request")
	switch action {
	case "volume":
		if vol := r.Form["vol"][0]; vol != "" {
			volInt, ok := strconv.Atoi(vol)
			if ok == nil {
				err = setVolume(volInt)
			}
		}
	case "seek":
		if time := r.Form["time"][0]; time != "" {
			timeInt, ok := strconv.Atoi(time)
			if ok == nil {
				err = seekPos(-1, timeInt)
				return
			}
		}
	case "update":
		err = mpdActionUpdate(r.FormValue("uri"))

	case "add_to_notes":
		status := getStatus()
		data := make([]string, 0)
		for key, val := range status.Current {
			data = append(data, fmt.Sprintf("%s: %s", key, val))
		}
		data = append(data, "\n-----------------\n\n")
		err = h.AppendToFile("mpd_notes.txt", strings.Join(data, "\n"))

	default:
		err = mpdAction(action)
	}

	switch action {
	case "playlist-clear":
		if err != nil {
			sess := app.GetSessionStore(w, r)
			sess.AddFlash(err.Error(), "error")
			app.SaveSession(w, r)
		}
		http.Redirect(w, r, app.GetNamedURL("mpd-playlist"), http.StatusFound)
	default:
		if err == nil {
			w.Write([]byte("OK"))
		} else {
			w.Write([]byte(err.Error()))
		}
	}
}

func mpdLogPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewSimpleDataPageCtx(w, r, "mpd", "mpd", "", localMenu)
	ctx.SetMenuActive("mpd-tools", "mpd-log")
	ctx.LocalMenu = localMenu
	ctx.Header1 = "Logs"

	if lines, err := h.ReadFile("/var/log/mpd/mpd.log", 25); err != nil {
		ctx.Data = err.Error()
	} else {
		ctx.Data = lines
	}
	app.RenderTemplateStd(w, ctx, "data.tmpl")
}

var infoHandlerCache = h.NewSimpleCache(1)

func infoHandler(w http.ResponseWriter, r *http.Request) {
	data := infoHandlerCache.Get(func() h.Value {
		status := getStatus()
		encoded, _ := json.Marshal(status)
		return encoded
	}).([]byte)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

const fakeResult = `{"Status":{"consume":"0","mixrampdb":"0.000000","mixrampdelay":"nan","nextsong":"1","nextsongid":"1","playlist":"2","playlistlength":"222","random":"0","repeat":"0","single":"0","song":"0","songid":"0","state":"stop","volume":"100","xfade":"0"},"Current":{"Album":"Café Del Mar - Classic I","Artist":"Jules Massenet","Date":"2002","Genre":"Baroque, Modern, Romantic, Classical","Id":"0","Last-Modified":"2013-09-27T06:14:59Z","Pos":"0","Time":"312","Title":"Meditation","Track":"01/12","file":"muzyka/mp3/cafe del mar/compilations/classics/2002, classic/01. jules massenet - meditation.mp3"},"Error":""}`

func songInfoHandler(w http.ResponseWriter, r *http.Request) {
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
		l.Warn("page.mpd filePageHandler: missing action ", r.Form)
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}
	uri := r.FormValue("uri")
	if uri == "" {
		l.Warn("page.mpd filePageHandler: missing uri ", r.Form)
		http.Error(w, "missing uri", http.StatusBadRequest)
		return
	}
	uri, _ = url.QueryUnescape(uri)
	err := mpdFileAction(uri, action)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoded, _ := json.Marshal("OK")
		w.Write(encoded)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func notesPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewSimpleDataPageCtx(w, r, "mpd", "mpd", "", localMenu)
	ctx.SetMenuActive("mpd-tools", "mpd-notes")
	ctx.LocalMenu = localMenu
	ctx.Header1 = "Notes"

	if lines, err := h.ReadFile("mpd_notes.txt", -1); err != nil {
		ctx.Data = err.Error()
	} else {
		ctx.Data = lines
	}
	app.RenderTemplateStd(w, ctx, "data.tmpl")
}
