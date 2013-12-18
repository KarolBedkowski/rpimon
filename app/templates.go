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

// RenderTemplate - render given template
func RenderTemplate(w http.ResponseWriter, ctx interface{}, name string, filenames ...string) {

	cacheLock.Lock()
	defer cacheLock.Unlock()

	templatePath := strings.Join(filenames, "|")
	ctemplate, ok := cacheItems[templatePath]
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
		cacheItems[templatePath] = ctemplate
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
