package common

import (
	"encoding/xml"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type Images struct{ open func(string) (fs.File, error) }

// Wrap dispatches media before ServeMux can clean dot segments into other routes.
func (h *Images) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/media" || strings.HasPrefix(r.URL.Path, mediaPrefix) {
			h.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Images) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name, ok := mediaName(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	want := imageTypes[strings.ToLower(path.Ext(name))]
	if want == "" {
		http.NotFound(w, r)
		return
	}
	f, err := h.open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	file, ok := f.(seekableFile)
	info, err := f.Stat()
	if err != nil || !ok || !info.Mode().IsRegular() || !matchesImage(file, want) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", want)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	http.ServeContent(w, r, path.Base(name), info.ModTime(), file)
}

var imageTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".gif": "image/gif", ".webp": "image/webp", ".ico": "image/x-icon",
	".svg": "image/svg+xml",
}

func matchesImage(file seekableFile, mime string) bool {
	if mime == "image/svg+xml" {
		d := xml.NewDecoder(io.LimitReader(file, 4096))
		for {
			token, err := d.Token()
			if err != nil {
				return false
			}
			if start, ok := token.(xml.StartElement); ok {
				return start.Name.Local == "svg" && start.Name.Space == "http://www.w3.org/2000/svg"
			}
		}
	}
	var buf [512]byte
	n, err := io.ReadFull(file, buf[:])
	return (err == nil || err == io.ErrUnexpectedEOF) && http.DetectContentType(buf[:n]) == mime
}
