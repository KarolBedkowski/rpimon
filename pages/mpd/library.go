package mpd

import (
	"encoding/json"
	"k.prv/rpimon/app"
	//	h "k.prv/rpimon/helpers"
	//"code.google.com/p/gompd/mpd"
	"github.com/turbowookie/gompd/mpd"
	l "k.prv/rpimon/helpers/logging"
	"net/http"
	"net/url"
	"strings"
)

func libraryPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := app.NewBasePageContext("Mpd", "mpd", w, r)
	ctx.LocalMenu = localMenu
	ctx.CurrentLocalMenuPos = "mpd-library"
	app.RenderTemplateStd(w, ctx, "mpd/library.tmpl")
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
		Artist      string
		Album       string
		AlbumArtist string
		Title       string
		Track       string
		Name        string
		Genre       string
		Date        string
		Composer    string
		Performer   string
		Comment     string
		Disc        string
		Filename    string
		Any         string
	}

	searchPageContext struct {
		*app.BasePageContext
		Form   searchForm
		Result []mpd.Attrs
	}
)

func (f *searchForm) getQueryString() (query string) {
	if f.Any != "" {
		return "any \"" + f.Any + "\""
	}
	q := make([]string, 0)
	if f.Artist != "" {
		q = append(q, "artist \""+f.Artist+"\"")
	}
	if f.Album != "" {
		q = append(q, "album \""+f.Album+"\"")
	}
	if f.AlbumArtist != "" {
		q = append(q, "albumartist \""+f.AlbumArtist+"\"")
	}
	if f.Title != "" {
		q = append(q, "title \""+f.Title+"\"")
	}
	if f.Track != "" {
		q = append(q, "track \""+f.Track+"\"")
	}
	if f.Name != "" {
		q = append(q, "name \""+f.Name+"\"")
	}
	if f.Genre != "" {
		q = append(q, "genre \""+f.Genre+"\"")
	}
	if f.Date != "" {
		q = append(q, "date \""+f.Date+"\"")
	}
	if f.Composer != "" {
		q = append(q, "composer \""+f.Composer+"\"")
	}
	if f.Performer != "" {
		q = append(q, "performer \""+f.Performer+"\"")
	}
	if f.Comment != "" {
		q = append(q, "comment \""+f.Comment+"\"")
	}
	if f.Disc != "" {
		q = append(q, "disc \""+f.Disc+"\"")
	}
	if f.Filename != "" {
		q = append(q, "filename \""+f.Filename+"\"")
	}
	return strings.Join(q, " ")
}

func searchPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := &searchPageContext{
		BasePageContext: app.NewBasePageContext("Mpd", "mpd", w, r),
	}
	ctx.LocalMenu = localMenu
	ctx.CurrentLocalMenuPos = "mpd-search"

	r.ParseForm()
	if err := decoder.Decode(ctx, r.Form); err != nil {
		l.Warn("Decode form error", err, r.Form)
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
