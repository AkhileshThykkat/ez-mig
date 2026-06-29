package runner

import (
	"strings"
	"testing"

	"github.com/AkhileshThykkat/ez-mig/internal/session"
)

func TestExecuteRejectsUnsupportedScheme(t *testing.T) {
	s := session.Session{
		Name:               "bad",
		DbURI:              "sqlite:///tmp/app.db",
		MigrationFilesPath: "/tmp/migrations",
	}

	err := Execute(s, ActionUp, nil)
	if err == nil {
		t.Fatal("expected error for unsupported scheme, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported database scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteRejectsInvalidURI(t *testing.T) {
	s := session.Session{
		Name:               "bad",
		DbURI:              "://missing-scheme",
		MigrationFilesPath: "/tmp/migrations",
	}

	if err := Execute(s, ActionUp, nil); err == nil {
		t.Fatal("expected error for invalid URI, got nil")
	}
}
