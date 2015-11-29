package mpd

import (
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	h "k.prv/rpimon/helpers"
	"net/http"
)

func mpdLogPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.NewDataPageCtx(w, r, "mpd")
	ctx.SetMenuActive("mpd-log")
	ctx.Header1 = "Logs"

	if lines, err := h.ReadFile("/var/log/mpd/mpd.log", 25); err != nil {
		ctx.Data = err.Error()
	} else {
		ctx.Data = lines
	}
	app.RenderTemplateStd(w, ctx, "data.tmpl")
}
