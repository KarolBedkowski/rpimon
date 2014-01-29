package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

// CreateRoutes for /storage
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(dfPageHandler, "admin")).Name("storage-index")
	subRouter.HandleFunc("/mount", app.VerifyPermission(mountPageHandler, "admin")).Name("storage-mount")
	subRouter.HandleFunc("/umount", app.VerifyPermission(umountPageHandler, "admin")).Name("storage-umount")
	subRouter.HandleFunc("/df", app.VerifyPermission(dfPageHandler, "admin")).Name("storage-df")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("storage-page")
}

func newPageCtx(w http.ResponseWriter, r *http.Request, localMenuPos string, data string) *app.SimpleDataPageCtx {
	ctx := app.NewSimpleDataPageCtx(w, r, "Storage", "storage", localMenuPos, createLocalMenu())
	ctx.CurrentLocalMenuPos = localMenuPos
	ctx.Data = data
	return ctx
}

var localMenu []*app.MenuItem

func createLocalMenu() []*app.MenuItem {
	if localMenu == nil {
		localMenu = []*app.MenuItem{app.NewMenuItemFromRoute("Disk Free", "storage-df").SetID("diskfree"),
			app.NewMenuItemFromRoute("Mount", "storage-mount").SetID("mount"),
			app.NewMenuItemFromRoute("Devices", "storage-page", "page", "devices").SetID("devices")}
	}
	return localMenu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, _ := vars["page"]
	ctx := newPageCtx(w, r, page, "")
	switch page {
	case "devices":
		ctx.Data = h.ReadFromCommand("lsblk")
	default:
		http.Redirect(w, r, app.GetNamedURL("storage-index"), 302)
		return
	}
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "data.tmpl", "flash.tmpl")
}

type mountPoint struct {
	Mpoint  string
	Device  string
	Type    string
	Options string
}

func mountPageHandler(w http.ResponseWriter, r *http.Request) {
	var ctx struct {
		*app.SimpleDataPageCtx
		CurrentPage string
		Data        string
		Mounted     []*mountPoint
	}
	ctx.SimpleDataPageCtx = app.NewSimpleDataPageCtx(w, r, "Storage", "storage", "storage", createLocalMenu())
	ctx.CurrentLocalMenuPos = "mount"
	ctx.Data = h.ReadFromCommand("sudo", "mount")
	ctx.Mounted = mountCmdToMountPoints(ctx.Data)
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "storage/storage.tmpl", "flash.tmpl")
}

func mountCmdToMountPoints(data string) (res []*mountPoint) {
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			break
		}
		fields := strings.Fields(line)
		res = append(res, &mountPoint{fields[2], fields[0], fields[4], fields[5]})
	}
	return
}

func umountPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	values := r.Form
	fs, ok := values["fs"]
	if ok && fs[0] != "" {
		data := h.ReadFromCommand("sudo", "umount", fs[0])
		if data != "" {
			ctx := newPageCtx(w, r, "mount", data)
			app.RenderTemplate(w, ctx, "base", "base.tmpl", "data.tmpl", "flash.tmpl")
			return
		}
	}
	sess := app.GetSessionStore(w, r)
	sess.AddFlash("Umounted "+fs[0], "success")
	sess.Save(r, w)

	http.Redirect(w, r, app.GetNamedURL("storage-mount"), 302)
}

func dfPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewSimpleDataPageCtx(w, r, "Storage", "storage", "storage", createLocalMenu())
	ctx.CurrentLocalMenuPos = "diskfree"
	ctx.TData = make([][]string, 0)
	ctx.THead = []string{"Filesystem", "Size", "Used", "Available", "Used %", "Mounted on"}
	lines := strings.Split(h.ReadFromCommand("df"), "\n")
	for _, line := range lines[1:] {
		if line != "" {
			ctx.TData = append(ctx.TData, strings.Fields(line))
		}
	}
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "data.tmpl", "flash.tmpl")
}
