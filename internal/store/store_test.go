package store

import (
	"database/sql"
	"testing"

	"github.com/AkhileshThykkat/ez-mig/internal/session"
	_ "modernc.org/sqlite"
)

// newTestDB swaps the package-level db for an isolated in-memory SQLite
// instance and bootstraps the schema. Each test gets a clean store.
func newTestDB(t *testing.T) {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	db = conn
	if err := createTable(); err != nil {
		t.Fatalf("create table: %v", err)
	}
}

func sample(name string) session.Session {
	return session.Session{
		Name:               name,
		DbURI:              "postgres://user:pass@localhost:5432/app?sslmode=disable",
		MigrationFilesPath: "/tmp/migrations",
	}
}

func TestCreateAndGetByName(t *testing.T) {
	newTestDB(t)

	in := sample("prod")
	if err := Create(in); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := GetByName("prod")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if got.Name != in.Name || got.DbURI != in.DbURI || got.MigrationFilesPath != in.MigrationFilesPath {
		t.Fatalf("round-trip mismatch: got %+v want %+v", got, in)
	}
}

func TestCreateDuplicateFails(t *testing.T) {
	newTestDB(t)

	if err := Create(sample("dup")); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if err := Create(sample("dup")); err == nil {
		t.Fatal("expected error on duplicate name, got nil")
	}
}

func TestGetByNameNotFound(t *testing.T) {
	newTestDB(t)

	if _, err := GetByName("ghost"); err == nil {
		t.Fatal("expected error for missing session, got nil")
	}
}

func TestUpdate(t *testing.T) {
	newTestDB(t)

	if err := Create(sample("stage")); err != nil {
		t.Fatalf("Create: %v", err)
	}

	upd := sample("stage")
	upd.DbURI = "mysql://root@localhost:3306/app"
	upd.MigrationFilesPath = "/srv/db/migrations"
	if err := Update(upd); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := GetByName("stage")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if got.DbURI != upd.DbURI || got.MigrationFilesPath != upd.MigrationFilesPath {
		t.Fatalf("update not persisted: got %+v", got)
	}
}

func TestUpdateMissingFails(t *testing.T) {
	newTestDB(t)

	if err := Update(sample("nope")); err == nil {
		t.Fatal("expected error updating missing session, got nil")
	}
}

func TestDelete(t *testing.T) {
	newTestDB(t)

	if err := Create(sample("temp")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := Delete("temp"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := GetByName("temp"); err == nil {
		t.Fatal("expected session gone after delete")
	}
}

func TestDeleteMissingFails(t *testing.T) {
	newTestDB(t)

	if err := Delete("nope"); err == nil {
		t.Fatal("expected error deleting missing session, got nil")
	}
}

func TestListOrderedAndEmpty(t *testing.T) {
	newTestDB(t)

	got, err := List()
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %d", len(got))
	}

	for _, n := range []string{"a", "b", "c"} {
		if err := Create(sample(n)); err != nil {
			t.Fatalf("Create %s: %v", n, err)
		}
	}

	got, err = List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(got))
	}
}
