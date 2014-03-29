package app

import (
	"html/template"
	"k.prv/rpimon/app/cfg"
	l "k.prv/rpimon/helpers/logging"
	res "k.prv/rpimon/resources/tmpl"
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

func getTemplate(name string, filenames ...string) (tmpl *template.Template) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	ctemplate, ok := cacheItems[name]
	if !ok {
		ctemplate = template.New(name).Funcs(funcMap)
		for _, name := range filenames {
			if !cfg.Configuration.Debug {
				if c, err := res.Asset(name); err == nil {
					ctemplate = template.Must(ctemplate.Parse(string(c)))
					continue
				} else {
					l.Error("RenderTemplate get template %s from box error: %s", name, err.Error())
				}
			}
			fullPath := filepath.Join(cfg.Configuration.TemplatesDir, name)
			if !fileExists(fullPath) {
				if cfg.Configuration.Debug {
					if c, err := res.Asset(name); err == nil {
						ctemplate = template.Must(ctemplate.Parse(string(c)))
						continue
					}
				}
				l.Error("RenderTemplate missing template: %s", fullPath)
			} else {
				ctemplate = template.Must(ctemplate.ParseFiles(fullPath))
			}
		}
		if ctemplate.Lookup("scripts") == nil {
			ctemplate, _ = ctemplate.Parse("{{define \"scripts\"}}{{end}}")
		}
		if ctemplate.Lookup("header") == nil {
			ctemplate, _ = ctemplate.Parse("{{define \"header\"}}{{end}}")
		}
		if ctemplate.Lookup("tabs") == nil {
			ctemplate, _ = ctemplate.Parse("{{define \"tabs\"}}{{end}}")
		}
		cacheItems[name] = ctemplate
	}
	return ctemplate
}

// RenderTemplate - render given templates.
func RenderTemplate(w http.ResponseWriter, ctx interface{}, name string, filenames ...string) {
	ctemplate := getTemplate(name, filenames...)
	if ctemplate == nil {
		return
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
