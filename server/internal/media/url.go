package media

import (
	"net/url"
	"path"
	"strings"
)

// ValidURL accepts hosted artwork or a file under the public media mount.
func ValidURL(value string) bool {
	if len(value) > 2048 || strings.ContainsAny(value, "\\\r\n\t ") {
		return false
	}
	u, err := url.Parse(value)
	if err != nil || u.User != nil || u.Opaque != "" {
		return false
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return u.Hostname() != ""
	}
	return u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/media/") && strings.HasPrefix(value, "/media/") && path.Clean(u.Path) == u.Path
}
