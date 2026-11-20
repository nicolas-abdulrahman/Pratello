// Package common provides behavior shared by multiple API domains.
package common

import (
	"errors"
	"io/fs"
	"net/url"
	"strings"
	"unicode"
)

var ErrImageURL = errors.New("invalid image URL")

const mediaPrefix = "/media/"

// ImageURL turns root-relative image references into public media URLs.
// External HTTP(S) references are returned unchanged for the client to load.
func ImageURL(value string) (string, error) {
	if len(value) > 2048 || strings.Contains(value, "\\") || strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return "", ErrImageURL
	}
	u, err := url.Parse(value)
	if err != nil || u.User != nil || u.Opaque != "" {
		return "", ErrImageURL
	}
	if (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != "" {
		return value, nil
	}
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || u.Scheme != "" || u.Host != "" {
		return "", ErrImageURL
	}
	if !strings.HasPrefix(u.Path, mediaPrefix) {
		// Keep the original escaping when adding the public mount prefix.
		value = "/media" + value
		u.Path = "/media" + u.Path
	} else if !strings.HasPrefix(value, mediaPrefix) {
		return "", ErrImageURL
	}
	if _, ok := mediaName(u.Path); !ok {
		return "", ErrImageURL
	}
	return value, nil
}

func mediaName(path string) (string, bool) {
	name, ok := strings.CutPrefix(path, mediaPrefix)
	if !ok || !fs.ValidPath(name) || strings.ContainsAny(name, "\\%") || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return "", false
	}
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			return "", false
		}
	}
	return name, true
}
