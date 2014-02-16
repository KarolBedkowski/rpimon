package mpd

import (
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
)

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
