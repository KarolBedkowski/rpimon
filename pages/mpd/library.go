package mpd

import (
	"encoding/json"
	"k.prv/rpimon/app"
	//	h "k.prv/rpimon/helpers"
	//	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"net/url"
	"strings"
)

type libraryPageCtx struct {
	*app.BasePageContext
	CurrentPage string
}

/*
type BreadcrumbItem struct {
	Title  string
	Href   string
	Active bool
}
*/

func libraryPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &libraryPageCtx{BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r)}
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentLocalMenuPos = "mpd-library"
	app.RenderTemplate(w, ctx, "base", "base.tmpl", "mpd/library.tmpl", "flash.tmpl")
}

func libraryActionHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	action := r.FormValue("a")
	switch action {
	case "add", "replace":
		if uriL, ok := r.Form["u"]; ok {
			uri := strings.TrimLeft(uriL[0], "/")
			uri, _ = url.QueryUnescape(uri)
			err := addFileToPlaylist(uri, action == "replace")
			if err == nil {
				w.Write([]byte("Added to playlist"))
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
	}
	http.Error(w, "Invalid request", http.StatusBadRequest)
}

type libraryContenet struct {
	Path  string     `json:"path"`
	Error string     `json:"error"`
	Items [][]string `json:"items"`
}

func libraryContentService(w http.ResponseWriter, r *http.Request) {
	path, _ := url.QueryUnescape(r.FormValue("p"))
	if len(path) > 0 {
		if path[0] != '/' {
			path = "/" + path
		}
		if path[len(path)-1] != '/' {
			path = path + "/"
		}
	}
	result := libraryContenet{Path: path}
	folders, files, err := getFiles(strings.Trim(path, "/"))
	if err != nil {
		result.Error = err.Error()
	} else {
		for _, folder := range folders {
			result.Items = append(result.Items, []string{"0", folder})
		}
		for _, file := range files {
			result.Items = append(result.Items, []string{"1", file})
		}
	}
	encoded, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(encoded)
}