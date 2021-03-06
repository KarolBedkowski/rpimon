package utils

import (
	"bufio"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/logging"
	"net/http"
	"os"
	"path/filepath"
)

// Module information
var Module *app.Module

func init() {
	Module = &app.Module{
		Name:        "files",
		Title:       "Files",
		Description: "File browser",
		Init:        initModule,
		GetMenu:     getMenu,
		Defaults: map[string]string{
			"config_file": "./browser.json",
		},
		Configurable:  true,
		AllPrivilages: []app.Privilege{{"files", "access to file browser"}},
	}
}

// CreateRoutes for /files
func initModule(parentRoute *mux.Route) bool {
	conf := Module.GetConfiguration()
	configFilename := conf["config_file"]
	if err := loadConfiguration(configFilename); err != nil {
		l.Warn("Files: failed load configuration: %s", err)
		return false
	}
	subRouter := parentRoute.Subrouter()
	subRouter.HandleFunc("/index",
		app.VerifyPermission(verifyAccess(mainPageHandler), "files")).Name(
		"files-index")
	subRouter.HandleFunc("/mkdir",
		app.VerifyPermission(verifyAccess(mkdirServHandler), "files")).Methods(
		"POST").Name("files-mkdir")
	subRouter.HandleFunc("/upload",
		app.VerifyPermission(verifyAccess(uploadPageHandler), "files")).Methods(
		"POST").Name("files-upload")
	subRouter.HandleFunc("/serv/dirs",
		app.VerifyPermission(dirServHandler, "files")).Name(
		"files-serv-dirs")
	subRouter.HandleFunc("/serv/files",
		app.VerifyPermission(filesServHandler, "files")).Name(
		"files-serv-files")
	subRouter.HandleFunc("/action",
		app.VerifyPermission(verifyAccess(actionHandler), "files")).Methods(
		"PUT").Name("files-file-action")
	return true
}

func getMenu(ctx *app.BaseCtx) (parentID string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "files") {
		return "", nil
	}
	return "", app.NewMenuItem("Files", app.GetNamedURL("files-index")).SetID("files").SetIcon("glyphicon glyphicon-hdd")
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, pctx *pathContext) {
	ctx := app.NewBaseCtx("Files", w, r)
	ctx.SetMenuActive("files")
	r.ParseForm()
	var relpath, abspath = ".", config.BaseDir
	if pctx != nil {
		relpath = pctx.relpath
		abspath = pctx.abspath
	}

	l.Debug("mainPageHandler: %v", relpath)
	isDirectory, err := isDir(abspath)
	if err != nil {
		app.Render404(w, r, "Error: "+err.Error())
		return
	}
	// Serve file
	if !isDirectory {
		l.Debug("files: serve file %s", abspath)
		w.Header().Set("Content-Disposition",
			"attachment; filename=\""+filepath.Base(abspath)+"\"")
		http.ServeFile(w, r, abspath)
		return
	}

	app.RenderTemplateStd(w, ctx, "files/browser.tmpl")
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request, pctx *pathContext) {
	if pctx == nil {
		app.Render404(w, r, "Error: missing path")
		return
	}
	var relpath, abspath = pctx.relpath, pctx.abspath
	if isDirectory, err := isDir(abspath); !isDirectory || err != nil {
		app.Render400(w, r, "Error: wrong path "+err.Error())
		return
	}

	f, handler, err := r.FormFile("upload")
	if err != nil {
		app.Render400(w, r, "Error: missing file "+err.Error())
		return
	}
	defer f.Close()

	fabspath := filepath.Join(abspath, handler.Filename)
	l.Debug("files: upload files %s", fabspath)
	file, err := os.Create(fabspath)
	if err != nil {
		app.Render400(w, r, "Error: error creating file "+err.Error())
		return
	}
	defer file.Close()
	out := bufio.NewWriter(file)
	io.Copy(out, f)
	out.Flush()
	http.Redirect(w, r,
		app.GetNamedURL("files-index")+h.BuildQuery("p", relpath),
		http.StatusFound)
}

func actionHandler(w http.ResponseWriter, r *http.Request, pctx *pathContext) {
	if pctx == nil {
		app.Render404(w, r, "Error: missing path")
		return
	}
	action, ok := h.GetParam(w, r, "action")
	if !ok {
		return
	}
	var relpath, abspath = pctx.relpath, pctx.abspath

	switch action {
	case "delete":
		l.Debug("Delete ", abspath)
		err := os.Remove(abspath)
		if err == nil {
			w.Write([]byte(abspath + " deleted"))
		} else {
			app.Render500(w, r, "Error: "+err.Error())
		}
		return
	case "move":
		l.Debug("Move %v ->", abspath)
		if dest := r.FormValue("d"); dest != "" {
			if adest, rdest, err := isPathValid(dest); err == nil {
				adest = filepath.Join(adest, filepath.Base(relpath))
				l.Debug("Move -> %v", adest)
				if rdest == relpath {
					return
				}
				err = os.Rename(abspath, adest)
				if err == nil {
					w.Write([]byte(relpath + " moved to " + rdest))
				} else {
					app.Render404(w, r, "Error: "+err.Error())
				}
				return
			}
		}
	}
	app.Render400(w, r, "Error: invalid action")
}

type dirInfo struct {
	ID       string          `json:"id"`
	Text     string          `json:"text"`
	Children interface{}     `json:"children"`
	State    map[string]bool `json:"state"`
}

func dirServHandler(w http.ResponseWriter, r *http.Request) {
	if config.BaseDir == "" {
		app.Render400(w, r, "Missing module configuration. Check browser.json")
		return
	}
	r.ParseForm()
	path, ok := h.GetParam(w, r, "id")
	if !ok {
		return
	}

	abspath, relpath, err := isPathValid(id2Dir(path))
	if err != nil {
		app.Render400(w, r, "Invalid id")
		return
	}

	var children []dirInfo
	if files, err := ioutil.ReadDir(abspath); err == nil {
		for _, file := range files {
			if file.Mode()&os.ModeSymlink == os.ModeSymlink {
				// Folow symlinks
				file, err = os.Stat(filepath.Join(relpath, file.Name()))
				if err != nil {
					continue
				}
			}
			if file.IsDir() {
				ipath := filepath.Join(relpath, file.Name())
				children = append(children, dirInfo{dir2ID(ipath), file.Name(), true, nil})
			}
		}
	}

	name := "Root"
	if relpath != "." {
		name = filepath.Base(relpath)
	}

	result := &dirInfo{dir2ID(relpath), name, children, nil}
	if relpath == "." {
		result.State = map[string]bool{"opened": true, "selected": true}
	}
	encoded, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}

func filesServHandler(w http.ResponseWriter, r *http.Request) {
	if config.BaseDir == "" {
		app.Render500(w, r, "Missing module configuration. Check browser.josm")
		return
	}
	r.ParseForm()
	path, ok := h.GetParam(w, r, "id")
	if !ok {
		return
	}
	abspath, relpath, err := isPathValid(path)
	if err != nil {
		app.Render400(w, r, "invalid id")
		return
	}
	var children [][]interface{}
	if path != "." {
		children = append(children, []interface{}{
			"folder",
			"..",
			"",
			"",
			filepath.Join(relpath, ".."),
		})
	}
	if files, err := ioutil.ReadDir(abspath); err == nil {
		for _, file := range files {
			if file.Mode()&os.ModeSymlink == os.ModeSymlink {
				// Folow symlinks
				file, err = os.Stat(filepath.Join(relpath, file.Name()))
				if err != nil {
					continue
				}
			}
			kind := "file"
			if file.IsDir() {
				kind = "dir"
			}
			ipath := filepath.Join(relpath, file.Name())
			finfo := []interface{}{
				kind,
				file.Name(),
				file.Size(),
				app.FormatDate(file.ModTime(), ""),
				ipath,
			}
			children = append(children, finfo)
		}
	}
	encoded, _ := json.Marshal(children)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}

func mkdirServHandler(w http.ResponseWriter, r *http.Request, pctx *pathContext) {
	if pctx == nil {
		app.Render400(w, r, "Error: missing path")
		return
	}
	if dirname, ok := h.GetParam(w, r, "name"); ok {
		dirpath := filepath.Join(pctx.abspath, dirname)
		l.Debug("files: create dir %s", dirpath)
		if err := os.MkdirAll(dirpath, os.ModePerm|0770); err != nil {
			app.Render500(w, r, "Error: creating directory: "+err.Error())
		}
		w.Write([]byte("OK"))
	}
}
