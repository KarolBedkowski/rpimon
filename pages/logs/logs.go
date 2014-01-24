package logs

import (
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /logs
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "admin")).Name("logs-index")
	subRouter.HandleFunc("/serv", app.VerifyPermission(servLogHandler, "admin")).Name("logs-serv")
	subRouter.HandleFunc("/{page}", app.VerifyPermission(mainPageHandler, "admin")).Name("logs-page")
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Data        string
	Files       []string
}

var localMenu []*app.MenuItem

func createLocalMenu() []*app.MenuItem {
	if localMenu == nil {
		localMenu = []*app.MenuItem{app.NewMenuItemFromRoute("Short", "logs-page", "page", "short").SetID("short"),
			app.NewMenuItemFromRoute("DMESG", "logs-page", "page", "dmesg").SetID("dmesg"),
			app.NewMenuItemFromRoute("Syslog", "logs-page", "page", "syslog").SetID("syslog")}
	}
	return localMenu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("logs", "logs", w, r)}
	ctx.LocalMenu = createLocalMenu()
	vars := mux.Vars(r)
	page, ok := vars["page"]
	if !ok {
		page = "short"
	}
	file := r.FormValue("file")

	if data, err := getLog(page, file); err == nil {
		ctx.Data = data
	} else {
		ctx.Data = err.Error()
	}
	switch page {
	case "syslog":
		ctx.Files = findFiles("syslog")
	}
	ctx.CurrentLocalMenuPos = page
	ctx.CurrentPage = page
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "logs.tmpl", "flash.tmpl")
}

func servLogHandler(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	page := r.FormValue("page")

	data, err := getLog(page, file)
	if err != nil {
		data = err.Error()
	}
	if strings.HasSuffix(file, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(data))
}

func getLogPath(filename string) (string, error) {
	abspath, err := filepath.Abs(filepath.Clean(filepath.Join("/var/log/", filename)))
	if !strings.HasPrefix(abspath, "/var/log/") {
		return "", errors.New("wrong path")
	}
	f, err := os.Open(abspath)
	if err != nil {
		return "", errors.New("not found")
	}
	defer f.Close()
	d, err1 := f.Stat()
	if err1 != nil || d.IsDir() {
		return "", errors.New("not found")
	}
	return abspath, err
}

func findFiles(prefix string) (result []string) {
	if files, err := ioutil.ReadDir("/var/log/"); err == nil {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), prefix) && !file.IsDir() {
				result = append(result, file.Name())
			}
		}
	}
	return
}

func getLog(page, file string) (string, error) {
	switch page {
	case "short":
		return h.ReadFromFileLastLines("/var/log/syslog", 20)
	case "dmesg":
		return h.ReadFromCommand("dmesg"), nil
	case "syslog":
		if file == "" {
			file = "syslog"
		}
		path, err := getLogPath(file)
		if err == nil {
			return h.ReadFromFileLastLines(path, 500)
		}
		return "", err
	}
	return "", errors.New("Invalid request")
}
