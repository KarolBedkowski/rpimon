package app

import (
	"html/template"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var cacheLock sync.Mutex
var cacheItems = map[string]*template.Template{}

var funcMap = template.FuncMap{
	"namedurl":   GetNamedURL,
	"formatDate": FormatDate,
}

func FormatDate(date time.Time, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return date.Format(format)
}

// RenderTemplate - render given template
func RenderTemplate(w http.ResponseWriter, ctx interface{}, name string, filenames ...string) {

	cacheLock.Lock()
	defer cacheLock.Unlock()

	templateKey := strings.Join(filenames, "|")
	ctemplate, ok := cacheItems[templateKey]
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
		ctemplate = template.New(templateKey).Funcs(funcMap)
		ctemplate = template.Must(ctemplate.ParseFiles(templates...))
		cacheItems[templateKey] = ctemplate
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
