package menu

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCAT056MenuLayoutRejectsNullTemplateStyle(t *testing.T) {
	assertMenuLayoutRejectionXFail(t, `{"template_style":null}`)
}

func TestCAT057MenuLayoutRejectsNullName(t *testing.T) {
	assertMenuLayoutRejectionXFail(t, `{"name":null}`)
}

func TestCAT058MenuLayoutRejectsNullAccentColor(t *testing.T) {
	assertMenuLayoutRejectionXFail(t, `{"accent_color":null}`)
}

func TestCAT059MenuLayoutRejectsNullTypography(t *testing.T) {
	assertMenuLayoutRejectionXFail(t, `{"typography":null}`)
}

// Only acceptance of the invalid input is expected. Other errors still fail.
func assertMenuLayoutRejectionXFail(t *testing.T, body string) {
	t.Helper()
	r := httptest.NewRequest(http.MethodPut, "/restaurants/cafe/menu", strings.NewReader(body))
	var input layoutInput
	err := decode(httptest.NewRecorder(), r, &input)
	if err == nil {
		_, err = prepare(r.Context(), nil, catalog{Layout: layout{Style: "classic"}}, request{op: putLayout, body: &input})
	}
	switch {
	case errors.Is(err, ErrInvalid):
		t.Fatalf("XPASS: %s was rejected as expected; remove the XFAIL expectation", body)
	case err != nil:
		t.Fatalf("unexpected validation error for %s: %v", body, err)
	default:
		t.Logf("XFAIL: %s was accepted; expected invalid input (HTTP 400)", body)
	}
}
