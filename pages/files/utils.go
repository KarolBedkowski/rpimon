package utils

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type pathContext struct {
	abspath string
	relpath string
}

type pathHandlerFunc func(w http.ResponseWriter, r *http.Request, pctx *pathContext)

func verifyAccess(h pathHandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pathD, ok := r.Form["p"]
		var ctx *pathContext
		if ok {
			abspath, relpath, err := isPathValid(pathD[0])
			if err != nil {
				http.Error(w, "Fobidden/wrong path "+err.Error(), http.StatusForbidden)
				return
			}
			if abspath != "" {
				ctx = &pathContext{abspath: abspath, relpath: relpath}
			}
		}
		h(w, r, ctx)
	})
}

func isPathValid(inputPath string) (abspath, relpath string, err error) {
	if inputPath == "" || inputPath == "#" {
		inputPath = "."
	}
	abspath, err = filepath.Abs(filepath.Clean(
		filepath.Join(config.BaseDir, inputPath)))
	if err != nil {
		return "", "", err
	}
	if !strings.HasPrefix(abspath, config.BaseDir) {
		return "", "", errors.New("wrong path")
	}
	if relpath, err = filepath.Rel(config.BaseDir, abspath); err != nil {
		return "", "", err
	}
	err = nil
	return
}

func isDir(filename string) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, errors.New("not found")
	}
	defer f.Close()
	d, err1 := f.Stat()
	if err1 != nil {
		return false, errors.New("not found")
	}
	return d.IsDir(), nil
}

func id2Dir(id string) string {
	if id == "dt--root" {
		return "."
	}
	path, _ := url.QueryUnescape(id)
	if strings.Index(path, "dt-") == 0 {
		return path[3:]
	}
	return id
}

func dir2ID(path string) string {
	if path == "." {
		return "dt--root"
	}
	return "dt-" + url.QueryEscape(path)
}
