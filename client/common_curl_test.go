package client

import (
	"net/http"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogCurlCommandMasksCredentials(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	saved := logger
	SetLogger(zap.New(core))
	defer SetLogger(saved)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/api/json", nil)
	req.SetBasicAuth("user", "secret-token")

	logCurlCommand(req)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(entries))
	}
	cmd := entries[0].ContextMap()["cmd"].(string)
	if want := "Authorization: <masked>"; !contains(cmd, want) {
		t.Fatalf("expected %q in %q", want, cmd)
	}
	if contains(cmd, "Basic") {
		t.Fatalf("credentials leaked: %q", cmd)
	}
	if user, token, ok := req.BasicAuth(); !ok || user != "user" || token != "secret-token" {
		t.Fatal("the request header must be restored")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
