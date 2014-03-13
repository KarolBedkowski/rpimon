package mpd

import (
	"encoding/json"
	"k.prv/rpimon/app"
	"k.prv/rpimon/app/context"
	"k.prv/rpimon/app/errors"
	//	h "k.prv/rpimon/helpers"
	//"code.google.com/p/gompd/mpd"
	"github.com/turbowookie/gompd/mpd"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"net/url"
	"strings"
)

func libraryPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	bctx.SetMenuActive("mpd-library")
	app.RenderTemplateStd(w, bctx, "mpd/library.tmpl")
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
				errors.Render400(w, r)
			}
			return
		}
	case "update":
		err := mpdActionUpdate(r.FormValue("uri"))
		if err != nil {
			errors.Render400(w, r)
		}
		return
	}
	errors.Render400(w, r)
}

func libraryServHandler(w http.ResponseWriter, r *http.Request) {
	path, _ := url.QueryUnescape(r.FormValue("p"))
	if len(path) > 0 {
		if path[0] != '/' {
			path = "/" + path
		}
		if path[len(path)-1] != '/' {
			path = path + "/"
		}
	}
	var result struct {
		Path  string     `json:"path"`
		Error string     `json:"error"`
		Items [][]string `json:"items"`
	}

	result.Path = path
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

type (
	searchForm struct {
		Field string
		Value string
	}

	searchPageContext struct {
		*context.BasePageContext
		Form   searchForm
		Result []mpd.Attrs
	}
)

func (f *searchForm) getQueryString() (query string) {
	f.Field = strings.TrimSpace(f.Field)
	f.Value = strings.TrimSpace(f.Value)
	if f.Field == "" || f.Value == "" {
		return ""
	}
	return f.Field + " \"" + f.Value + "\""
}

func searchPageHandler(w http.ResponseWriter, r *http.Request, bctx *context.BasePageContext) {
	ctx := &searchPageContext{BasePageContext: bctx}
	ctx.SetMenuActive("mpd-search")

	r.ParseForm()
	if err := decoder.Decode(ctx, r.Form); err != nil {
		l.Debug("pages.mpd.library.searchPageHandler decode form error %s %#v", err.Error(), r.Form)
	}

	if r.Method == "POST" {
		if query := ctx.Form.getQueryString(); query != "" {
			result, err := find(query)
			if err == nil {
				ctx.Result = result
				for _, item := range ctx.Result {
					if item["Artist"] == "" {
						item["Artist"] = item["AlbumArtist"]
					}
					if item["Title"] == "" {
						item["Title"] = item["file"]
					}
				}
			} else {
				l.Error("searchPageContext error: %s", err.Error())
			}
		}
	}
	app.RenderTemplateStd(w, ctx, "mpd/search.tmpl")
}
