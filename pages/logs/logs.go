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
	"strconv"
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
			app.NewMenuItemFromRoute("Syslog", "logs-page", "page", "syslog").SetID("syslog"),
			app.NewMenuItemFromRoute("Samba", "logs-page", "page", "samba").SetID("samba"),
		}
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

	switch page {
	case "syslog":
		ctx.Files = findFiles("/var/log/", "syslog")
	case "samba":
		ctx.Files = findFiles("/var/log/samba", "log")
	}
	file := r.FormValue("file")
	if file == "" && ctx.Files != nil && len(ctx.Files) > 0 {
		file = ctx.Files[0]
	}
	if data, err := getLog(page, file, 100); err == nil {
		ctx.Data = data
	} else {
		ctx.Data = err.Error()
	}
	ctx.CurrentLocalMenuPos = page
	ctx.CurrentPage = page
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "logs.tmpl", "flash.tmpl")
}

func servLogHandler(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	page := r.FormValue("page")

	linelimit := 100
	if lines := r.FormValue("lines"); lines != "" {
		if limit, err := strconv.Atoi(lines); err == nil {
			linelimit = limit
		}
	}

	data, err := getLog(page, file, linelimit)
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
		return "", errors.New("wrong path " + filename)
	}
	f, err := os.Open(abspath)
	if err != nil {
		return "", errors.New("not found " + abspath)
	}
	defer f.Close()
	d, err1 := f.Stat()
	if err1 != nil || d.IsDir() {
		return "", errors.New("not found " + abspath)
	}
	return abspath, err
}

func findFiles(dir, prefix string) (result []string) {
	if files, err := ioutil.ReadDir(dir); err == nil {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), prefix) && !file.IsDir() {
				result = append(result, file.Name())
			}
		}
	}
	return
}

func getLog(page, file string, lines int) (result string, err error) {
	if strings.HasSuffix(file, ".gz") {
		lines = -1
	}
	switch page {
	case "short":
		result, err = h.ReadFromFileLastLines("/var/log/syslog", 20)
	case "dmesg":
		result, err = h.ReadFromCommand("dmesg"), nil
	case "syslog":
		path, err := getLogPath(file)
		if err != nil {
			return "", err
		}
		result, err = h.ReadFromFileLastLines(path, lines)
	case "samba":
		path, err := getLogPath("samba/" + file)
		if err != nil {
			return "", err
		}
		result, err = h.ReadFromFileLastLines(path, lines)
	default:
		return "", errors.New("Invalid request")
	}
	if result == "" {
		result = "<EMPTY FILE>"
	}
	return result, nil
}
