package notepad

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"io/ioutil"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var decoder = schema.NewDecoder()

// ErrInvalidFilename error - invalid filename
var ErrInvalidFilename = errors.New("invalid filename")

// Module information
var Module *app.Module

func init() {
	Module = &app.Module{
		Name:        "notepad",
		Title:       "Notepad",
		Description: "",
		Init:        initModule,
		GetMenu:     getMenu,
		Defaults: map[string]string{
			"dir": "./notepad/",
		},
		Configurable:  true,
		AllPrivilages: []app.Privilege{{"notepad", "access to notepad"}},
	}
}

// CreateRoutes for /mpd
func initModule(parentRoute *mux.Route) bool {
	subRouter := parentRoute.Subrouter()
	// Main page
	subRouter.HandleFunc("/", app.SecContext(mainPageHandler, "Notepad", "notepad")).Name("notepad-index")
	subRouter.HandleFunc("/{note}", app.VerifyPermission(notePageHandler, "notepad")).Name("notepad-note")
	subRouter.HandleFunc("/{note}/delete", app.VerifyPermission(noteDeleteHandler, "notepad")).Name("notepad-delete")
	subRouter.HandleFunc("/{note}/download", app.VerifyPermission(noteDownloadHandler, "notepad")).Name("notepad-download")
	return true
}

func getMenu(ctx *app.BaseCtx) (parentID string, menu *app.MenuItem) {
	if ctx.CurrentUser == "" || !app.CheckPermission(ctx.CurrentUserPerms, "notepad") {
		return "", nil
	}
	menu = app.NewMenuItem("Notepad", app.GetNamedURL("notepad-index")).SetID("notepad-index").SetIcon("glyphicon glyphicon-paperclip")
	return "", menu
}

// NoteStuct keep information about one note
type NoteStuct struct {
	Filename string
	Content  string
}

// Validate note
func (n *NoteStuct) Validate() (errors []string) {
	if n.Filename == "" {
		errors = append(errors, "missing filename")
	}
	return
}

type mainPageContext struct {
	*app.BaseCtx
	NoteList []*NoteStuct
}

func mainPageHandler(r *http.Request, bctx *app.BaseCtx) {
	ctx := &mainPageContext{BaseCtx: bctx}
	ctx.SetMenuActive("notepad-index")
	ctx.NoteList = findFiles()
	ctx.RenderStd(ctx, "notepad/index.tmpl")
}

type notePageContext struct {
	*app.BaseCtx
	Note *NoteStuct
	New  bool
}

func notePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename, _ := vars["note"]
	if filename == "" {
		app.Render400(w, r, "Invalid request: missing filename")
		return
	}
	switch r.Method {
	case "GET":
		// display note
		ctx := &notePageContext{BaseCtx: app.NewBaseCtx("Notepad", w, r)}
		if note, err := getNote(filename); err == nil {
			ctx.Note = note
		} else {
			ctx.Note = &NoteStuct{Filename: filename}
			ctx.New = true
		}
		ctx.SetMenuActive("notepad-index")
		app.RenderTemplateStd(w, ctx, "notepad/note.tmpl")
		return
	case "POST":
		// save note
		r.ParseForm()
		note := new(NoteStuct)
		decoder.Decode(note, r.Form)
		sess := app.GetSessionStore(w, r)
		if err := SaveNote(filename, note.Content); err == nil {
			sess.AddFlash("Note saved", "success")
		} else {
			sess.AddFlash(err.Error(), "error")
		}
		app.SaveSession(w, r)
		http.Redirect(w, r, app.GetNamedURL("notepad-index"), http.StatusFound)
		return
	case "DELETE":
		// delete note
		filepath, _ := getFilepath(filename)
		if filepath == "" {
			app.Render400(w, r, "Invalid request: invalid filename")
			return
		}
	}
	app.Render400(w, r)
}

func noteDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename, _ := vars["note"]
	if filename == "" {
		app.Render400(w, r, "Invalid Request: missing filename")
		return
	}
	sess := app.GetSessionStore(w, r)
	if err := DeleteNote(filename); err == nil {
		sess.AddFlash("Note deleted", "success")
	} else {
		sess.AddFlash(err.Error(), "error")
	}
	app.SaveSession(w, r)
	http.Redirect(w, r, app.GetNamedURL("notepad-index"), http.StatusFound)
}

func noteDownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename, _ := vars["note"]
	if filename == "" {
		app.Render400(w, r, "Invalid Request: missing filename")
		return
	}
	if filepath, ok := getFilepath(filename); ok {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		http.ServeFile(w, r, filepath)
	} else {
		app.Render404(w, r, "File not found")
	}
}

func getNotepadDir() (string, bool) {
	conf := Module.GetConfiguration()
	notepadDir, err := filepath.Abs(filepath.Clean(conf["dir"]))
	if err != nil {
		l.Error("notepad.findFiles %s, %s", conf["dir"], err.Error())
		return "", false
	}
	return notepadDir, true
}

func findFiles() (result []*NoteStuct) {
	if notepadDir, ok := getNotepadDir(); ok {
		if files, err := ioutil.ReadDir(notepadDir); err == nil {
			for _, file := range files {
				if !file.IsDir() {
					result = append(result, &NoteStuct{
						Filename: file.Name(),
					})
				}
			}
		}
	}
	return
}

func getFilepath(filename string) (path string, ok bool) {
	if notepadDir, okdir := getNotepadDir(); okdir {
		abspath, err := filepath.Abs(filepath.Clean(filepath.Join(notepadDir, filename)))
		if err != nil {
			l.Error("notepad.getFilepath %s, %s", filename, err.Error())
			return
		}
		if !strings.HasPrefix(abspath, notepadDir) {
			l.Error("notepad.getFilepath %s, bad prefix: %s; notepadDir=%s", filename, abspath, notepadDir)
			return
		}
		return abspath, true
	}
	return
}

func getNote(filename string) (note *NoteStuct, err error) {
	var content []byte
	filepath, ok := getFilepath(filename)
	if !ok {
		return nil, ErrInvalidFilename
	}
	if content, err = ioutil.ReadFile(filepath); err == nil {
		note = &NoteStuct{
			Filename: filename,
			Content:  string(content),
		}
	}
	return
}

// SaveNote write content to new or truncated file
func SaveNote(filename string, content string) error {
	filepath, ok := getFilepath(filename)
	if !ok {
		return ErrInvalidFilename
	}
	return ioutil.WriteFile(filepath, []byte(content), 0600)
}

// AppendToNote append data to existing file; create if not exists.
func AppendToNote(filename string, content string) error {
	filepath, ok := getFilepath(filename)
	if !ok {
		return ErrInvalidFilename
	}
	return h.AppendToFile(filepath, content)
}

// DeleteNote remove file with given name
func DeleteNote(filename string) error {
	filepath, ok := getFilepath(filename)
	if !ok {
		return ErrInvalidFilename
	}
	return os.Remove(filepath)
}
