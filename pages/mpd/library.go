package mpd

import (
	"encoding/json"
	"k.prv/rpimon/app"
	//	h "k.prv/rpimon/helpers"
	//	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type libraryPageCtx struct {
	*app.BasePageContext
	CurrentPage string
	Path        string
	Files       []string
	Folders     []string
	Error       string
	Breadcrumb  []BreadcrumbItem
}

type BreadcrumbItem struct {
	Title  string
	Href   string
	Active bool
}

func libraryPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &libraryPageCtx{BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r)}
	ctx.LocalMenu = createLocalMenu()
	ctx.CurrentLocalMenuPos = "mpd-library"
	ctx.Path = ""
	r.ParseForm()
	if path, ok := r.Form["p"]; ok {
		ctx.Path, _ = url.QueryUnescape(strings.TrimLeft(path[0], "/"))
	}

	ctx.Breadcrumb = append(ctx.Breadcrumb, BreadcrumbItem{"[Library]", "", false})
	if ctx.Path != "" && ctx.Path != "." {
		prevPath := ""
		for idx, pElem := range strings.Split(ctx.Path, "/") {
			ctx.Breadcrumb[idx].Active = true
			prevPath = filepath.Join(prevPath, pElem)
			ctx.Breadcrumb = append(ctx.Breadcrumb, BreadcrumbItem{pElem, prevPath, false})
		}
	}
	var err error
	ctx.Folders, ctx.Files, err = getFiles(ctx.Path)
	if err != nil {
		ctx.Error = err.Error()
	}
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
