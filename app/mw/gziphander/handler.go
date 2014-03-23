// Handle gziped static files
// Based on: https://github.com/joaodasilva/go-gzip-file-server

package gziphandler

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type gzipFileHandler struct {
	fs           http.FileSystem
	cacheControl bool
}

func (h *gzipFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/") {
		r.URL.Path = "/" + r.URL.Path
	}
	serveFile(w, r, h.fs, path.Clean(r.URL.Path), true, h.cacheControl)
}

// FileServer - net.http.FileServer + serving <file>.gz when exists instead of requested
// file.
// When cacheControl - add header Cache-Control to response.
func FileServer(root http.FileSystem, cacheControl bool) http.Handler {
	return &gzipFileHandler{root, cacheControl}
}

// ServeFile - net.http.ServeFile but first try to serve gzip file when exists <file>.gz
func ServeFile(w http.ResponseWriter, r *http.Request, name string, cacheControl bool) {
	dir, file := filepath.Split(name)
	serveFile(w, r, http.Dir(dir), file, false, cacheControl)
}

func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem,
	name string, redirect bool, cacheControl bool) {

	// try to serve gziped file; ignore request for gz files
	if !strings.HasSuffix(strings.ToLower(name), ".gz") && supportsGzip(r) {
		if file, stat := open(fs, name+".gz"); file != nil {
			defer file.Close()
			name = stat.Name()
			name = name[:len(name)-3]
			setContentType(w, name, file)
			w.Header().Set("Content-Encoding", "gzip")
			if cacheControl {
				w.Header().Set("Cache-Control", "must_revalidate, private, max-age=604800")
			}
			http.ServeContent(w, r, name, stat.ModTime(), file)
			return
		}
	}

	// serve requested file
	if file, stat := open(fs, name); file != nil {
		defer file.Close()
		if cacheControl {
			w.Header().Set("Cache-Control", "must_revalidate, private, max-age=604800")
		}
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
		return
	}

	http.NotFound(w, r)
}

func supportsGzip(r *http.Request) bool {
	for _, encodings := range r.Header["Accept-Encoding"] {
		for _, encoding := range strings.Split(encodings, ",") {
			if encoding == "gzip" {
				return true
			}
		}
	}
	return false
}

func setContentType(w http.ResponseWriter, name string, file http.File) {
	t := mime.TypeByExtension(filepath.Ext(name))
	if t == "" {
		var buffer [512]byte
		n, _ := io.ReadFull(file, buffer[:])
		t = http.DetectContentType(buffer[:n])
		if _, err := file.Seek(0, os.SEEK_SET); err != nil {
			http.Error(w, "Can't seek", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", t)
}

// open file and return File and FileInfo; ignore directories.
func open(fs http.FileSystem, name string) (file http.File, stat os.FileInfo) {
	var err error
	file, err = fs.Open(name)
	if err != nil {
		return
	}
	stat, err = file.Stat()
	if err != nil || stat.IsDir() { // ignore dirs
		file.Close()
		file = nil
	}
	return
}
