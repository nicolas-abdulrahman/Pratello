package user

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProfileDoesNotSerializeCredentials(t *testing.T) {
	data, err := json.Marshal(Account{ID: 42, Password: "private-hash", TokenVersion: 9})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private-hash") || strings.Contains(strings.ToLower(string(data)), "tokenversion") {
		t.Fatalf("profile exposed credentials: %s", data)
	}
}
