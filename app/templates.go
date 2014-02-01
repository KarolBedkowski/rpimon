package app

import (
	"html/template"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var cacheLock sync.Mutex
var cacheItems = map[string]*template.Template{}

var funcMap = template.FuncMap{
	"namedurl":   GetNamedURL,
	"formatDate": FormatDate,
}

// FormatDate in template
func FormatDate(date time.Time, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return date.Format(format)
}

// MainTemplateName contains name of main section in template (main template)
const MainTemplateName = "base"

// RenderTemplate - render given templates.
func RenderTemplate(w http.ResponseWriter, ctx interface{}, name string, filenames ...string) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	ctemplate, ok := cacheItems[name]
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
		ctemplate = template.New(name).Funcs(funcMap)
		ctemplate = template.Must(ctemplate.ParseFiles(templates...))
		if ctemplate.Lookup("scripts") == nil {
			ctemplate, _ = ctemplate.Parse("{{define \"scripts\"}}{{end}}")
		}
		cacheItems[name] = ctemplate
	}
	err := ctemplate.ExecuteTemplate(w, MainTemplateName, ctx)
	if err != nil {
		l.Error("RenderTemplate execution failed: %s on %s (%s)", err,
			name, filenames)
	}
}

// StdTemplates contains list of templates included when rendering by RenderTemplateStd
var StdTemplates = []string{"base.tmpl", "flash.tmpl"}

// RenderTemplateStd render given templates + StdTemplates.
// Main section in template must be named 'base'.
// First template file name is used as template name.
func RenderTemplateStd(w http.ResponseWriter, ctx interface{}, filenames ...string) {
	filenames = append(filenames, StdTemplates...)
	l.Debug("RenderTemplateStd; %v", filenames)
	RenderTemplate(w, ctx, filenames[0], filenames...)
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
