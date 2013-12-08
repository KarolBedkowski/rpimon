package app

import (
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

var cacheLock sync.Mutex
var cacheItems = map[string]*template.Template{}

func RenderTemplate(w http.ResponseWriter, ctx interface{}, name string, filenames ...string) {

	cacheLock.Lock()
	defer cacheLock.Unlock()

	template_path := strings.Join(filenames, "|")
	ctemplate, ok := cacheItems[template_path]
	if !ok || Configuration.Debug {
		templates := []string{}
		for _, filename := range filenames {
			fullPath := filepath.Join(Configuration.TemplatesDir, filename)
			if !fileExists(fullPath) {
				l.Error("RenderTemplate missing template: %s", fullPath)
				return
			}
			templates = append(templates, fullPath)
		}
		ctemplate = template.Must(template.ParseFiles(templates...))
		cacheItems[template_path] = ctemplate
	}
	err := ctemplate.ExecuteTemplate(w, name, ctx)
	if err != nil {
		l.Error("RenderTemplate execution failed: %s on %s (%s)", err,
			name, filenames)
	}
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			l.Error(name, " does not exist.")
		}
		return false
	}
	return true
}
