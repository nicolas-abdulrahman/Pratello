package common

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMediaHTTPContainmentAndImageResponses(t *testing.T) {
	work := t.TempDir()
	media := filepath.Join(work, "media")
	if err := os.MkdirAll(filepath.Join(media, "menu"), 0755); err != nil {
		t.Fatal(err)
	}
	write := func(path string, data []byte) {
		t.Helper()
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var picture bytes.Buffer
	if err := png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 3, 3))); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(media, "menu", "burger.png"), picture.Bytes())
	write(filepath.Join(media, "menu", "fake.png"), []byte("private configuration disguised as an image"))
	write(filepath.Join(media, "secret.txt"), []byte("private media data"))
	write(filepath.Join(media, ".hidden.png"), picture.Bytes())
	write(filepath.Join(media, "menu", "logo.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	outside := filepath.Join(work, "outside.png")
	write(outside, picture.Bytes())
	if err := os.Symlink(outside, filepath.Join(media, "escape.png")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(work, filepath.Join(media, "outside")); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /outside.png", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })
	h := ImageFiles(media).Wrap(mux)
	request := func(method, path string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, path, nil))
		return w
	}
	w := request("GET", "/media/menu/burger.png?v=1")
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" || !bytes.Equal(w.Body.Bytes(), picture.Bytes()) {
		t.Fatalf("valid image did not round-trip: %d %v", w.Code, w.Header())
	}
	if w := request("HEAD", "/media/menu/burger.png"); w.Code != 200 || w.Body.Len() != 0 {
		t.Fatalf("invalid HEAD response: %d", w.Code)
	}
	if w := request("GET", "/media/menu/logo.svg"); w.Code != 200 || !strings.Contains(w.Header().Get("Content-Security-Policy"), "sandbox") || w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("SVG missing isolation headers: %d %v", w.Code, w.Header())
	}
	for _, path := range []string{
		"/media", "/media/", "/media/menu/", "/media/menu/missing.png", "/media/menu/fake.png", "/media/secret.txt", "/media/.hidden.png",
		"/media/../outside.png", "/media/%2e%2e/outside.png", "/media/%252e%252e/outside.png", "/media/menu%5c..%5csecret.txt",
		"/media/escape.png", "/media/outside/outside.png",
	} {
		if w := request("GET", path); w.Code != 404 || w.Header().Get("Location") != "" {
			t.Errorf("%s was not rejected directly: %d %v", path, w.Code, w.Header())
		}
	}
	if w := request("POST", "/media/menu/burger.png"); w.Code != 405 || w.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("media accepts unexpected method: %d", w.Code)
	}
	if w := request("GET", "/outside.png"); w.Code != 204 {
		t.Fatalf("normal API routing changed: %d", w.Code)
	}
}
