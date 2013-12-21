package utils

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	//	h "k.prv/rpimon/helpers"
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /files
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/",
		app.VerifyPermission(verifyAccess(mainPageHandler), "files")).Name("files-index")
	subRouter.HandleFunc("/mkdir",
		app.VerifyPermission(verifyAccess(mkdirPageHandler), "files")).Methods("POST")
	subRouter.HandleFunc("/upload",
		app.VerifyPermission(verifyAccess(uploadPageHandler), "files")).Methods("POST")
}

type pageCtx struct {
	*app.BasePageContext
	CurrentPage   string
	Configuration configuration
	Files         []os.FileInfo
	Path          string
}

func (ctx pageCtx) GetFullPath(path string) string {
	return filepath.Join(ctx.Path, path)
}

func newPageCtx(w http.ResponseWriter, r *http.Request) *pageCtx {
	ctx := &pageCtx{BasePageContext: app.NewBasePageContext("Files", w, r)}
	ctx.CurrentMainMenuPos = "/files/"
	ctx.Configuration = config
	return ctx
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newPageCtx(w, r)
	r.ParseForm()
	var relpath, abspath = ".", config.BaseDir

	if relpathd, ok := r.Form["REL_PATH"]; ok {
		relpath = relpathd[0]
	}
	if abspathd, ok := r.Form["ABS_PATH"]; ok {
		abspath = abspathd[0]
	}

	ctx.Path = relpath

	isDirectory, err := isDir(abspath)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusNotFound)
		return
	}
	// Serve file
	if !isDirectory {
		l.Debug("files: serve file %s", abspath)
		http.ServeFile(w, r, abspath)
		return
	}
	// show dir
	l.Debug("files: serve dir %s", abspath)
	if files, err := ioutil.ReadDir(abspath); err == nil {
		ctx.Files = files
	} else {
		http.Error(w, "Error "+err.Error(), http.StatusBadRequest)
	}

	app.RenderTemplate(w, ctx, "base", "base.tmpl", "files/browser.tmpl", "flash.tmpl")
}

func mkdirPageHandler(w http.ResponseWriter, r *http.Request) {
	var relpath, abspath = "", ""
	if relpathd, ok := r.Form["REL_PATH"]; ok {
		relpath = relpathd[0]
	}
	if abspathd, ok := r.Form["ABS_PATH"]; ok {
		abspath = abspathd[0]
	}
	if relpath == "" || abspath == "" {
		http.Error(w, "Error: missing path ", http.StatusBadRequest)
		return
	}
	dirnamel, ok := r.Form["name"]
	if !ok {
		http.Error(w, "Error: missing dir name", http.StatusNotFound)
	}
	dirname := dirnamel[0]
	dirpath := filepath.Join(abspath, dirname)
	l.Debug("files: create dir %s", dirpath)
	if err := os.MkdirAll(dirpath, os.ModePerm|0770); err != nil {
		http.Error(w, "Error: creating directory "+err.Error(),
			http.StatusNotFound)
	}
	http.Redirect(w, r, app.GetNamedURL("files-index")+
		app.PairsToQuery("p", relpath), http.StatusFound)
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	var relpath, abspath = "", ""
	if relpathd, ok := r.Form["REL_PATH"]; ok {
		relpath = relpathd[0]
	}
	if abspathd, ok := r.Form["ABS_PATH"]; ok {
		abspath = abspathd[0]
	}
	if relpath == "" || abspath == "" {
		http.Error(w, "Error: missing path ", http.StatusBadRequest)
		return
	}
	if isDirectory, err := isDir(abspath); !isDirectory || err != nil {
		http.Error(w, "Error: wrong path "+err.Error(), http.StatusBadRequest)
		return
	}

	f, handler, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, "missing file "+err.Error(), http.StatusBadRequest)
		return
	}
	defer f.Close()

	fabspath := filepath.Join(abspath, handler.Filename)
	l.Debug("files: upload files %s", fabspath)
	file, err := os.Create(fabspath)
	if err != nil {
		http.Error(w, "error creating file: "+err.Error(), http.StatusForbidden)
		return
	}
	defer file.Close()
	out := bufio.NewWriter(file)
	io.Copy(out, f)
	out.Flush()
	http.Redirect(w, r,
		app.GetNamedURL("files-index")+app.PairsToQuery("p", relpath),
		http.StatusFound)
}

func isPathValid(inputPath string) (abspath, relpath string, err error) {
	abspath, err = filepath.Abs(filepath.Clean(
		filepath.Join(config.BaseDir, inputPath)))
	if err != nil {
		return "", "", err
	}
	if !strings.HasPrefix(abspath, config.BaseDir) {
		return "", "", errors.New("Wrong path")
	}
	if relpath, err = filepath.Rel(config.BaseDir, abspath); err != nil {
		return "", "", err
	}
	err = nil
	return
}

func verifyAccess(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pathD, ok := r.Form["p"]
		if ok {
			abspath, relpath, err := isPathValid(pathD[0])
			if err != nil {
				http.Error(w, "Fobidden/wrong path "+err.Error(), http.StatusForbidden)
				return
			}
			r.Form["ABS_PATH"] = []string{abspath}
			r.Form["REL_PATH"] = []string{relpath}
		}
		h(w, r)
	})
}

func isDir(filename string) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, errors.New("not found")
	}
	defer f.Close()
	d, err1 := f.Stat()
	if err1 != nil {
		return false, errors.New("not found")
	}
	return d.IsDir(), nil
}
