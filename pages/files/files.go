package utils

import (
	"github.com/gorilla/mux"
	"k.prv/rpimon/app"
	//	h "k.prv/rpimon/helpers"
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	//	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var subRouter *mux.Router

// CreateRoutes for /files
func CreateRoutes(parentRoute *mux.Route) {
	subRouter = parentRoute.Subrouter()
	subRouter.HandleFunc("/", app.VerifyLogged(mainPageHandler)).Name("files-index")
	subRouter.HandleFunc("/mkdir", app.VerifyLogged(mkdirPageHandler))
	subRouter.HandleFunc("/upload", app.VerifyLogged(uploadPageHandler))
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
	if pathD, ok := r.Form["p"]; ok {
		var err error
		abspath, relpath, err = isPathValid(pathD[0])
		if err != nil {
			http.Error(w, "Fobidden/wrong path "+err.Error(),
				http.StatusForbidden)
			return
		}
	}
	ctx.Path = relpath

	f, err := os.Open(abspath)
	if err != nil {
		http.Error(w, "Error: Not found ", http.StatusNotFound)
		return
	}
	defer f.Close()
	d, err1 := f.Stat()
	if err1 != nil {
		http.Error(w, "Error: Not found ", http.StatusNotFound)
		return
	}

	if !d.IsDir() {
		http.ServeFile(w, r, abspath)
		return
	}

	if files, err := ioutil.ReadDir(abspath); err == nil {
		ctx.Files = files
	} else {
		http.Error(w, "Error "+err.Error(), http.StatusBadRequest)
	}

	app.RenderTemplate(w, ctx, "base", "base.tmpl", "files/browser.tmpl", "flash.tmpl")
}

func mkdirPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	dirnamel, ok := r.Form["name"]
	if !ok {
		http.Error(w, "Error: missing dir name", http.StatusNotFound)
	}
	dirname := dirnamel[0]
	pathD, ok := r.Form["p"]
	if !ok {
		http.Error(w, "Error: missing path ", http.StatusNotFound)
	}
	abspath, relpath, err := isPathValid(filepath.Join(pathD[0], dirname))
	if err != nil {
		http.Error(w, "Fobidden/wrong path "+err.Error(),
			http.StatusForbidden)
		return
	}
	err = os.MkdirAll(abspath, os.ModePerm|0770)
	if err != nil {
		http.Error(w, "Error: creating directory "+err.Error(),
			http.StatusNotFound)
	}
	http.Redirect(w, r, app.GetNamedURL("files-index")+
		app.PairsToQuery("p", relpath), http.StatusFound)
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pathD, ok := r.Form["p"]
	if !ok {
		http.Error(w, "Error: missing path ", http.StatusNotFound)
	}

	dirname := pathD[0]
	f, handler, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, "missing file "+err.Error(), http.StatusBadRequest)
		return
	}
	defer f.Close()

	fname := handler.Filename
	fabspath, _, err := isPathValid(filepath.Join(dirname, fname))
	if err != nil {
		http.Error(w, "Fobidden/wrong path "+err.Error(), http.StatusForbidden)
		return
	}
	file, err := os.Create(fabspath)
	if err != nil {
		http.Error(w, "error creating file: "+err.Error(), http.StatusForbidden)
		return
	}
	defer file.Close()
	out := bufio.NewWriter(file)
	io.Copy(out, f)
	out.Flush()
	http.Redirect(w, r, app.GetNamedURL("files-index")+
		app.PairsToQuery("p", dirname), http.StatusFound)
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
