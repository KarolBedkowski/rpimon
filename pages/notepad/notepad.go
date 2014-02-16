package notepad

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"io/ioutil"
	"k.prv/rpimon/app"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var decoder = schema.NewDecoder()
var ErrInvalidFilename = errors.New("invalid filename")

// CreateRoutes for /mpd
func CreateRoutes(parentRoute *mux.Route) {
	subRouter := parentRoute.Subrouter()
	// Main page
	subRouter.HandleFunc("/", app.VerifyPermission(mainPageHandler, "notepad")).Name("notepad-index")
	subRouter.HandleFunc("/{note}", app.VerifyPermission(notePageHandler, "notepad")).Name("notepad-note")
	subRouter.HandleFunc("/{note}/delete", app.VerifyPermission(noteDeleteHandler, "notepad")).Name("notepad-delete")
}

type NoteStuct struct {
	Filename string
	Content  string
}

func (n *NoteStuct) Validate() (errors []string) {
	if n.Filename == "" {
		errors = append(errors, "missing filename")
	}
	return
}

type mainPageContext struct {
	*app.BasePageContext
	NoteList []*NoteStuct
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &mainPageContext{BasePageContext: app.NewBasePageContext("Notepad", "notepad", w, r)}
	ctx.SetMenuActive("notepad-index", "tools")
	ctx.NoteList = findFiles()
	app.RenderTemplateStd(w, ctx, "notepad/index.tmpl")
}

type notePageContext struct {
	*app.BasePageContext
	Note *NoteStuct
	New  bool
}

func notePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename, _ := vars["note"]
	if filename == "" {
		http.Error(w, "missing filename", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case "GET":
		// display note
		ctx := &notePageContext{BasePageContext: app.NewBasePageContext("Notepad", "notepad", w, r)}
		if note, err := getNote(filename); err == nil {
			ctx.Note = note
		} else {
			ctx.Note = &NoteStuct{Filename: filename}
			ctx.New = true
		}
		ctx.SetMenuActive("notepad-index", "tools")
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
			http.Error(w, "invalid filename", http.StatusBadRequest)
			return
		}
	}
	http.Error(w, "bad request", http.StatusBadRequest)
}

func noteDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename, _ := vars["note"]
	if filename == "" {
		http.Error(w, "missing filename", http.StatusBadRequest)
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

func findFiles() (result []*NoteStuct) {
	if app.Configuration.Notepad == "" {
		return
	}
	if files, err := ioutil.ReadDir(app.Configuration.Notepad); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				result = append(result, &NoteStuct{
					Filename: file.Name(),
				})
			}
		}
	}
	return
}

func getFilepath(filename string) (path string, ok bool) {
	abspath, err := filepath.Abs(filepath.Clean(filepath.Join(app.Configuration.Notepad, filename)))
	if err != nil {
		l.Error("notepad.getFilepath %s, %s", filename, err.Error())
		return
	}
	if !strings.HasPrefix(abspath, app.Configuration.Notepad) {
		l.Error("notepad.getFilepath %s, bad prefix: %s", filename, abspath)
		return
	}
	return abspath, true
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
