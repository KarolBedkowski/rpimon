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
	p := r.URL.Path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
		r.URL.Path = p
	}
	serveFile(w, r, h.fs, path.Clean(p), true, h.cacheControl)
}

// Similar to net.http.FileServer, but serves <file>.gz instead of <file> if
// it exists, has a later modification time, and the request supports gzip
// encoding. Also serves the .gz file if the original doesn't exist.
func FileServer(root http.FileSystem, cacheControl bool) http.Handler {
	return &gzipFileHandler{root, cacheControl}
}

// Similar to net.http.ServeFile, but serves <file>.gz instead of <file> if
// it exists, has a later modification time, and the request supports gzip
// encoding. Also serves the .gz file if the original doesn't exist.
func ServeFile(w http.ResponseWriter, r *http.Request, name string, cacheControl bool) {
	dir, file := filepath.Split(name)
	serveFile(w, r, http.Dir(dir), file, false, cacheControl)
}

func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem,
	name string, redirect bool, cacheControl bool) {

	if !strings.HasSuffix(strings.ToLower(name), ".gz") && supportsGzip(r) {
		file, stat := open(fs, name+".gz")
		if file != nil && !stat.IsDir() {
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

	file, stat := open(fs, name)
	if file != nil {
		defer file.Close()
		if !stat.IsDir() {
			if cacheControl {
				w.Header().Set("Cache-Control", "must_revalidate, private, max-age=604800")
			}
			http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
		}
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

func open(fs http.FileSystem, name string) (http.File, os.FileInfo) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, nil
	}
	s, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil
	}
	return f, s
}
