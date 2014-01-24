package users

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /storage
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("storage-index")
	subRouter.HandleFunc("/mount", app.VerifyPermission(mountPageHandler, "admin")).Name("storage-mount")
	subRouter.HandleFunc("/umount", app.VerifyPermission(umountPageHandler, "admin")).Name("storage-umount")
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
		localMenu = []*app.MenuItem{app.NewMenuItemFromRoute("Disk Free", "storage-page", "page", "diskfree").SetID("diskfree"),
			app.NewMenuItemFromRoute("Mount", "storage-mount").SetID("mount"),
			app.NewMenuItemFromRoute("Devices", "storage-page", "page", "devices").SetID("devices")}
	}
	return localMenu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "diskfree"
	}
	ctx := newPageCtx(w, r, page, "")
	switch page {
	case "diskfree":
		ctx.Data = h.ReadFromCommand("df", "-h")
	case "mount":
		ctx.Data = h.ReadFromCommand("sudo", "mount")
	case "devices":
		ctx.Data = h.ReadFromCommand("lsblk")
	default:
		http.Redirect(w, r, app.GetNamedURL("storage-index"), 302)
		return
	}
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
}

type mountPoint struct {
	Mpoint  string
	Device  string
	Type    string
	Options string
}

type mountPageCtx struct {
	*app.SimpleDataPageCtx
	CurrentPage string
	Data        string
	Mounted     []*mountPoint
}

func mountPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &mountPageCtx{SimpleDataPageCtx: app.NewSimpleDataPageCtx(w, r,
		"Storage", "storage", "storage", createLocalMenu())}
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
			app.RenderTemplate(w, ctx, "base", "base.tmpl", "log.tmpl", "flash.tmpl")
			return
		}
	}
	sess := app.GetSessionStore(w, r)
	sess.AddFlash("Umounted " + fs[0])
	sess.Save(r, w)

	http.Redirect(w, r, app.GetNamedURL("storage-mount"), 302)
}
