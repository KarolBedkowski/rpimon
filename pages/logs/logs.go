package logs

import (
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
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
	Logs        []string
	LogsGroup   logsGroup
	LogsDef     logsDef
}

var localMenu []*app.MenuItem

func createLocalMenu() []*app.MenuItem {
	if localMenu == nil {
		for _, group := range config.Groups {
			localMenu = append(localMenu,
				app.NewMenuItemFromRoute(group.Name, "logs-page", "page", group.Name).SetID(group.Name))
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
		page = config.Groups[0].Name
	}

	logname := r.FormValue("log")

	logs, group, err := findGroup(page, logname)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx.LogsGroup = group
	ctx.LogsDef = logs

	for _, logsdef := range group.Logs {
		ctx.Logs = append(ctx.Logs, logsdef.Name)
	}

	file := r.FormValue("file")
	ctx.Files = findFiles(logs)

	if file == "" {
		if logs.Filename != "" {
			file = logs.Filename
		} else {
			if ctx.Files != nil && len(ctx.Files) > 0 {
				file = ctx.Files[0]
			}
		}
	}
	if data, err := getLog(logs, file, 100); err == nil {
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
	logname := r.FormValue("log")
	page := r.FormValue("page")

	logs, _, err := findGroup(page, logname)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	linelimit := 100
	if lines := r.FormValue("lines"); lines != "" {
		if limit, err := strconv.Atoi(lines); err == nil {
			linelimit = limit
		}
	}

	data, err := getLog(logs, file, linelimit)
	if err != nil {
		data = err.Error()
	}
	if strings.HasSuffix(file, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(data))
}

func findFiles(log logsDef) (result []string) {
	if log.Dir == "" {
		return
	}
	if files, err := ioutil.ReadDir(log.Dir); err == nil {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), log.Prefix) && !file.IsDir() {
				result = append(result, file.Name())
			}
		}
	}
	return
}

func getLog(log logsDef, file string, lines int) (result string, err error) {
	l.Debug("getLog %v %s %d", log, file, lines)
	if strings.HasSuffix(file, ".gz") {
		lines = -1
	} else if log.Limit > 0 {
		lines = log.Limit
	} else if lines == 0 {
		lines = 50
	}

	if log.Command != "" {
		result = h.ReadFromCommand(log.Command)
	} else {
		var logpath string
		if log.Dir == "" {
			logpath = log.Filename
		} else {
			logpath, err = filepath.Abs(filepath.Clean(filepath.Join(log.Dir, file)))
			if err != nil {
				return "", err
			}
			if !strings.HasPrefix(logpath, log.Dir) {
				return "", errors.New("Invalid path")
			}
		}
		result, err = h.ReadFromFileLastLines(logpath, lines)
	}
	if result == "" && err == nil {
		result = "<EMPTY FILE>"
	}
	return result, nil
}
