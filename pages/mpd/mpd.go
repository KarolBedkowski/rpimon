package mpd

import (
	"code.google.com/p/gompd/mpd"
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
)

var subRouter *mux.Router

func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("mpd-index")
	subRouter.HandleFunc("/action/{action}", app.VerifyLogged(actionPageHandler))
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Status      *mpdStatus
}

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	ctx.CurrentMainMenuPos = "/mpd/"
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	data := newPageCtx(w, r)
	data.Status = getStatus()
	l.Print(data.Status)
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/index.tmpl", "flash.tmpl")
}

func actionPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	action, ok := vars["action"]
	if !ok || action == "" {
		l.Warn("page.mpd actionPageHandler: missing action ", vars)
		mainPageHandler(w, r)
		return
	}
	data := newPageCtx(w, r)
	data.Status = mpdAction(action)
	app.RenderTemplate(w, data, "base", "base.tmpl", "mpd/index.tmpl", "flash.tmpl")
}

type playlistPageCtx struct {
	*app.BasePageContext
	CurrentPage   string
	Playlist      []mpd.Attrs
	CurrentSongId string
	Error         error
}

func newPlaylistPageCtx(w http.ResponseWriter, r *http.Request) *playlistPageCtx {
	ctx := &playlistPageCtx{BasePageContext: app.NewBasePageContext("Mpd", w, r)}
	ctx.CurrentMainMenuPos = "/mpd/"
	return ctx
}

func playlistPageHandler(w http.ResponseWriter, r *http.Request) {

}
