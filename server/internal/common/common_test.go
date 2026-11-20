package common

import (
	"errors"
	"testing"
)

func TestImageURL(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"/joao_burger/menu/burger.png", "/media/joao_burger/menu/burger.png"},
		{"/media/account/bob.png", "/media/account/bob.png"},
		{"/menu/two%20burgers.png?v=2", "/media/menu/two%20burgers.png?v=2"},
		{"https://cdn.example.test/menu.png?signature=abc%2Fdef", "https://cdn.example.test/menu.png?signature=abc%2Fdef"},
		{"http://cdn.example.test/menu.png", "http://cdn.example.test/menu.png"},
		{"/", ""}, {"/media/", ""}, {"/media/../secret.png", ""},
		{"/%2e%2e/secret.png", ""}, {"/%252e%252e/secret.png", ""}, {"/menu%5csecret.png", ""},
		{"/media/.git/file.png", ""}, {"//cdn.example.test/menu.png", ""},
		{"menu.png", ""}, {"file:///private.png", ""}, {"javascript:alert(1)", ""},
		{"https://name:password@cdn.example.test/menu.png", ""},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ImageURL(tc.input)
			if tc.want == "" {
				if !errors.Is(err, ErrImageURL) {
					t.Fatalf("accepted invalid reference: %q (%v)", got, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q (%v), want %q", got, err, tc.want)
			}
		})
	}
}
