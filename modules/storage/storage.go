package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/session"
	h "k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

// Module information
var Module = &context.Module{
	Name:          "storage",
	Title:         "Storage",
	Description:   "",
	AllPrivilages: nil,
	Init:          initModule,
	GetMenu:       getMenu,
	Configurable:  true,
}

// CreateRoutes for /storage
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(dfPageHandler, "admin")).Name("storage-index")
	subRouter.HandleFunc("/mount", app.VerifyPermission(mountPageHandler, "admin")).Name("storage-mount")
	subRouter.HandleFunc("/umount", app.VerifyPermission(umountPageHandler, "admin")).Name("storage-umount")
	subRouter.HandleFunc("/df", app.VerifyPermission(dfPageHandler, "admin")).Name("storage-df")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("storage-page")
	return true
}

func getMenu(ctx *context.BasePageContext) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "admin") {
		return "", nil
	}
	menu = context.NewMenuItem("Storage", "").SetID("storage").SetIcon("glyphicon glyphicon-hdd")
	menu.AddChild(app.NewMenuItemFromRoute("Disk Free", "storage-df").SetID("diskfree"),
		app.NewMenuItemFromRoute("Mount", "storage-mount").SetID("mount"),
		app.NewMenuItemFromRoute("Devices", "storage-page", "page", "devices").SetID("devices"),
	)
	return "", menu
}

func newPageCtx(w http.ResponseWriter, r *http.Request, localMenuPos string, data string) *context.SimpleDataPageCtx {
	ctx := context.NewSimpleDataPageCtx(w, r, "Storage")
	ctx.SetMenuActive(localMenuPos)
	ctx.Data = data
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, _ := vars["page"]
	ctx := newPageCtx(w, r, page, "")
	switch page {
	case "devices":
		sec := r.FormValue("sec")
		if sec == "" {
			sec = "devices"
		}
		ctx.Header1 = "Storage"
		switch sec {
		case "devices":
			ctx.Header2 = "Devices"
			ctx.Data = h.ReadCommand("lsblk", "-a")
		case "fdisk":
			ctx.Header2 = "Fdisk"
			ctx.Data = h.ReadCommand("fdisk", "-l")
		}
		ctx.Tabs = []*context.MenuItem{
			app.NewMenuItemFromRoute("Devices", "storage-page", "page", page).AddQuery("?sec=devices").SetActve(sec == "devices"),
			app.NewMenuItemFromRoute("Fdisk", "storage-page", "page", page).AddQuery("?sec=fdisk").SetActve(sec == "fdisk"),
		}
	default:
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	app.RenderTemplateStd(w, ctx, "data.tmpl", "tabs.tmpl")
}

type (
	mountPoint struct {
		Mpoint  string
		Device  string
		Type    string
		Options string
	}

	mountPageContext struct {
		*context.SimpleDataPageCtx
		Data    string
		Mounted []*mountPoint
	}
)

func mountPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &mountPageContext{
		SimpleDataPageCtx: context.NewSimpleDataPageCtx(w, r, "Storage"),
	}
	ctx.SetMenuActive("mount")
	ctx.Header1 = "Storage"
	ctx.Header2 = "Mount"
	ctx.Data = h.ReadCommand("mount")
	ctx.Mounted = mountCmdToMountPoints(ctx.Data)
	app.RenderTemplateStd(w, ctx, "storage/storage.tmpl")
}

func mountCmdToMountPoints(data string) (res []*mountPoint) {
	for _, line := range strings.Split(data, "\n") {
		if line != "" {
			fields := strings.Fields(line)
			fields[5] = strings.Replace(fields[5], ",", ", ", -1)
			res = append(res, &mountPoint{fields[2], fields[0], fields[4], fields[5]})
		}
	}
	return
}

func umountPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fs := r.FormValue("fs")
	if fs == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	data := h.ReadCommand("sudo", "umount", fs)
	sess := session.GetSessionStore(w, r)
	if data != "" {
		sess.AddFlash("Umount "+fs+" error: "+data, "error")
	} else {
		sess.AddFlash("Umounted "+fs, "success")
	}
	sess.Save(r, w)
	http.Redirect(w, r, app.GetNamedURL("storage-mount"), http.StatusFound)
}

func dfPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.NewSimpleDataPageCtx(w, r, "Storage")
	ctx.SetMenuActive("diskfree")
	ctx.Header1 = "Storage"
	ctx.Header2 = "diskfree"
	ctx.TData = make([][]string, 0)
	ctx.THead = []string{"Filesystem", "Size", "Used", "Available", "Used %", "Mounted on"}
	lines := strings.Split(h.ReadCommand("df"), "\n")
	for _, line := range lines[1:] {
		if line != "" {
			ctx.TData = append(ctx.TData, strings.Fields(line))
		}
	}
	app.RenderTemplateStd(w, ctx, "data.tmpl")
}
