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
	subRouter.HandleFunc("/smart", app.VerifyPermission(smartPageHandler, "admin")).Name("storage-smart")
	subRouter.HandleFunc("/serv/smart", app.VerifyPermission(servSmartHandler, "admin")).Name("storage-serv-smart")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("storage-page")
}
func buildLocalMenu() (localMenu []*app.MenuItem) {
	return []*app.MenuItem{app.NewMenuItemFromRoute("Disk Free", "storage-df").SetID("diskfree"),
		app.NewMenuItemFromRoute("Mount", "storage-mount").SetID("mount"),
		app.NewMenuItemFromRoute("Devices", "storage-page", "page", "devices").SetID("devices"),
		app.NewMenuItemFromRoute("SMART", "storage-smart").SetID("smart"),
	}
}

func newPageCtx(w http.ResponseWriter, r *http.Request, localMenuPos string, data string) *app.SimpleDataPageCtx {
	ctx := app.NewSimpleDataPageCtx(w, r, "Storage", "storage", buildLocalMenu())
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
		ctx.Tabs = []*app.MenuItem{
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
		*app.SimpleDataPageCtx
		Data    string
		Mounted []*mountPoint
	}
)

func mountPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &mountPageContext{
		SimpleDataPageCtx: app.NewSimpleDataPageCtx(w, r, "Storage", "storage", buildLocalMenu()),
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
	sess := app.GetSessionStore(w, r)
	if data != "" {
		sess.AddFlash("Umount "+fs+" error: "+data, "error")
	} else {
		sess.AddFlash("Umounted "+fs, "success")
	}
	sess.Save(r, w)
	http.Redirect(w, r, app.GetNamedURL("storage-mount"), http.StatusFound)
}

func dfPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewSimpleDataPageCtx(w, r, "Storage", "storage", buildLocalMenu())
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

type smartPageContext struct {
	*app.SimpleDataPageCtx
	Devices []string
}

func smartPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &smartPageContext{SimpleDataPageCtx: app.NewSimpleDataPageCtx(w, r, "Storage - SMART", "storage",
		buildLocalMenu())}
	ctx.SetMenuActive("smart")
	for _, line := range strings.Split(h.ReadCommand("lsblk", "-r"), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "disk") {
			fields := strings.Fields(line)
			ctx.Devices = append(ctx.Devices, fields[0])
		}
	}
	app.RenderTemplateStd(w, ctx, "storage/smart.tmpl")
}

func servSmartHandler(w http.ResponseWriter, r *http.Request) {
	dev := r.FormValue("dev")
	smart := h.ReadCommand("sudo", "smartctl", "--all", "/dev/"+dev)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(smart))
}
