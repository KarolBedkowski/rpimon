package worker

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/session"
	"k.prv/rpimon/model"
	//h "k.prv/rpimon/helpers"
	//l "k.prv/rpimon/logging"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var decoder = schema.NewDecoder()

// Module information
var Module *context.Module
var dispatcher *Dispatcher

func init() {
	Module = &context.Module{
		Name:        "worker",
		Title:       "Worker",
		Description: "Run some commands in background",
		Init:        initModule,
		GetMenu:     getMenu,
		Defaults: map[string]string{
			"Logs_Dir":         "./worker-logs",
			"Default_Dir":      "./",
			"Parallel_workers": "2",
		},
		Configurable:  true,
		AllPrivilages: []context.Privilege{{"worker", "allow to run task"}},
	}

}

// CreateRoutes for /mpd
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	// active tasks
	subRouter.HandleFunc("/", context.SecContext(mainPageHandler, "Worker", "worker")).Name("worker-index")
	// new task
	subRouter.HandleFunc("/new", app.VerifyPermission(taskPageHandler, "worker")).Name("worker-new-task")
	// show
	subRouter.HandleFunc("/{idx:[0-9+]}", app.VerifyPermission(taskPageHandler, "worker")).Name("worker-task")
	// logfile
	subRouter.HandleFunc("/log/{name}", app.VerifyPermission(taskLogPageHandler, "worker")).Name("worker-task-log")

	conf := Module.GetConfiguration()
	workers, err := strconv.Atoi(conf["Parallel_workers"])
	if err != nil {
		workers = 2
	}
	dispatcher = NewDispatcher(workers)
	dispatcher.Run()

	go deleteOldLogs()
	return true
}

func getMenu(ctx *context.BaseCtx) (parentID string, menu *context.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "worker") {
		return "", nil
	}
	menu = context.NewMenuItem("Worker", app.GetNamedURL("worker-index")).SetID("worker-index").SetIcon("glyphicon glyphicon-flash")
	return "", menu
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BaseCtx) {
	ctx := &struct {
		*context.BaseCtx
		Tasks []*model.Task
	}{
		BaseCtx: bctx,
		Tasks:           model.GetTasks(),
	}
	ctx.SetMenuActive("worker-index")
	app.RenderTemplateStd(w, ctx, "worker/index.tmpl")
}

type taskPageContext struct {
	*context.BaseCtx
	Task *model.Task
}

func taskPageHandler(w http.ResponseWriter, r *http.Request) {

	ctx := &taskPageContext{
		BaseCtx: context.NewBaseCtx("Task", w, r),
		Task:            &model.Task{},
	}
	conf := Module.GetConfiguration()
	ctx.Task.Dir = conf["Default_Dir"]

	vars := mux.Vars(r)
	if idxs, ok := vars["idx"]; ok {
		if idx, err := strconv.Atoi(idxs); err == nil {
			if tsk := model.GetTask(idx); tsk != nil {
				ctx.Task = tsk
			}
		}
	}

	ctx.SetMenuActive("worker-index")
	if r.Method == "POST" {
		r.ParseForm()
		decoder.Decode(ctx.Task, r.Form)
		sess := session.GetSessionStore(w, r)
		success := false
		if err := ctx.Task.Validate(); err == nil {
			if ctx.Task.Multi && ctx.Task.Params != "" {
				params := strings.Split(ctx.Task.Params, "\n")
				for _, param := range params {
					task := ctx.Task.Clone()
					task.Params = strings.TrimSpace(param)
					task.Multi = false
					model.SaveTask(task)
					dispatcher.Add(Job{task})
				}
				sess.AddFlash(string(len(params))+" tasks created", "success")
				success = true
			} else {
				model.SaveTask(ctx.Task)
				dispatcher.Add(Job{ctx.Task})
				sess.AddFlash("Task created", "success")
				success = true
			}
		} else {
			sess.AddFlash(err.Error(), "error")
		}
		session.SaveSession(w, r)
		if success {
			http.Redirect(w, r, app.GetNamedURL("worker-index"), http.StatusFound)
			return
		}
	}
	app.RenderTemplateStd(w, ctx, "worker/task.tmpl")
}

func taskLogPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename, _ := vars["name"]
	if filename == "" {
		app.Render400(w, r, "Invalid Request: missing filename")
		return
	}

	dir := getLogsDir()
	abspath, err := filepath.Abs(path.Join(dir, filepath.Clean(filepath.Join(filename))))
	if err == nil && strings.HasPrefix(abspath, dir) {
		//w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		w.Header().Set("Content-Type", "text/plain")
		http.ServeFile(w, r, abspath)
		return
	}
	app.Render404(w, r, "File not found")
}
