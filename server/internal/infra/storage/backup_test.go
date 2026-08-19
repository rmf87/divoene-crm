package storage

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// withTempDir points BACKUP_DIR at a fresh temp dir for the duration of the test.
func withTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("BACKUP_DIR", dir)
	return dir
}

func newSQLite(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO t (v) VALUES ('hello')`); err != nil {
		t.Fatalf("insert: %v", err)
	}
	return db
}

func readSQLiteValue(t *testing.T, path string) string {
	t.Helper()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	var v string
	if err := db.QueryRow(`SELECT v FROM t LIMIT 1`).Scan(&v); err != nil {
		t.Fatalf("query: %v", err)
	}
	return v
}

func TestCreateAndListSnapshots(t *testing.T) {
	withTempDir(t)
	db := newSQLite(t, filepath.Join(t.TempDir(), "src.sqlite3"))

	path, err := CreateSnapshot(db)
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot file missing: %v", err)
	}

	files, err := ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}
	if files[0].LastLoaded {
		t.Errorf("fresh snapshot should not be marked as last loaded")
	}
}

func TestMarkLoadedAndDelete(t *testing.T) {
	withTempDir(t)
	db := newSQLite(t, filepath.Join(t.TempDir(), "src.sqlite3"))
	path, err := CreateSnapshot(db)
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	name := filepath.Base(path)

	if err := MarkLoaded(name); err != nil {
		t.Fatalf("MarkLoaded: %v", err)
	}

	files, _ := ListSnapshots()
	if len(files) != 1 || !files[0].LastLoaded {
		t.Fatalf("expected last_loaded on the snapshot, got %+v", files)
	}

	meta := LoadMeta()
	if meta.LastLoaded != name || len(meta.History) == 0 {
		t.Fatalf("meta not persisted: %+v", meta)
	}
	if meta.History[len(meta.History)-1].Action != "restore" {
		t.Errorf("expected restore action in history")
	}

	if err := DeleteSnapshot(name); err != nil {
		t.Fatalf("DeleteSnapshot: %v", err)
	}
	meta = LoadMeta()
	if meta.LastLoaded != "" {
		t.Errorf("last_loaded should be cleared after delete, got %q", meta.LastLoaded)
	}
}

func TestSaveUploadedSnapshot(t *testing.T) {
	withTempDir(t)
	srcPath := filepath.Join(t.TempDir(), "valid.sqlite3")
	src := newSQLite(t, srcPath)
	src.Close()

	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	name, err := SaveUploadedSnapshot(bytes.NewReader(data), "restore-20260701.sqlite3")
	if err != nil {
		t.Fatalf("SaveUploadedSnapshot: %v", err)
	}
	if name != "restore-20260701.sqlite3" {
		t.Errorf("got name %q, want original", name)
	}

	if _, err := SaveUploadedSnapshot(bytes.NewReader([]byte("not a sqlite db")), "bad.db"); err == nil {
		t.Error("expected error for invalid sqlite header")
	}
}

func TestSaveUploadedZip(t *testing.T) {
	withTempDir(t)

	validPath := filepath.Join(t.TempDir(), "one.sqlite3")
	db := newSQLite(t, validPath)
	db.Close()
	validData, _ := os.ReadFile(validPath)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, _ := zw.Create("folder/nested/one.sqlite3")
	fw.Write(validData)
	ignored, _ := zw.Create("readme.txt")
	ignored.Write([]byte("ignored"))
	zw.Close()

	names, err := SaveUploadedZip(bytes.NewReader(buf.Bytes()), "bundle.zip")
	if err != nil {
		t.Fatalf("SaveUploadedZip: %v", err)
	}
	if len(names) != 1 {
		t.Fatalf("got %d names, want 1", len(names))
	}
	if got := readSQLiteValue(t, filepath.Join(Dir(), names[0])); got != "hello" {
		t.Errorf("uploaded content = %q, want hello", got)
	}

	if _, err := SaveUploadedZip(bytes.NewReader([]byte("definitely not zip")), "bundle.zip"); err == nil {
		t.Error("expected error for invalid zip")
	}
}

func TestRestoreSnapshot(t *testing.T) {
	withTempDir(t)

	srcPath := filepath.Join(t.TempDir(), "snap.sqlite3")
	src := newSQLite(t, srcPath)
	src.Close()

	name, err := SaveUploadedSnapshot(mustRead(t, srcPath), "snap.sqlite3")
	if err != nil {
		t.Fatalf("SaveUploadedSnapshot: %v", err)
	}

	destPath := filepath.Join(Dir(), "live.sqlite3")
	if err := RestoreSnapshot(name, destPath); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}
	if got := readSQLiteValue(t, destPath); got != "hello" {
		t.Errorf("restored content = %q, want hello", got)
	}
}

func mustRead(t *testing.T, path string) *bytes.Reader {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return bytes.NewReader(data)
}

func ExampleDir() {
	fmt.Println(Dir() == defaultBackupDir || os.Getenv("BACKUP_DIR") != "")
	// Output: true
}
