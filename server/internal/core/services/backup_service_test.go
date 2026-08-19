package services

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rmf87/divoene/internal/core/domain"
)

type stubSnapshotStore struct {
	names []string
	dir   string
	meta  domain.BackupMeta
}

func (s *stubSnapshotStore) Create() (string, error) {
	n := "divoene-20060102-150405.sqlite3"
	s.names = append(s.names, n)
	path := filepath.Join(s.dir, n)
	if err := os.WriteFile(path, []byte("snapshot"), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *stubSnapshotStore) List() ([]domain.SnapshotInfo, error) {
	var out []domain.SnapshotInfo
	for _, n := range s.names {
		out = append(out, domain.SnapshotInfo{Name: n, Size: 10, LastLoaded: n == s.meta.LastLoaded})
	}
	return out, nil
}

func (s *stubSnapshotStore) Meta() domain.BackupMeta { return s.meta }

func (s *stubSnapshotStore) Path(name string) (string, error) {
	for _, n := range s.names {
		if n == name {
			return filepath.Join(s.dir, n), nil
		}
	}
	return "", errors.New("backup: file not found: " + name)
}

func (s *stubSnapshotStore) Delete(name string) error {
	for i, n := range s.names {
		if n == name {
			s.names = append(s.names[:i], s.names[i+1:]...)
			return nil
		}
	}
	return errors.New("backup: file not found: " + name)
}

func (s *stubSnapshotStore) SaveFromReader(r io.Reader, filename string) (string, error) {
	name := strings.ReplaceAll(filename, ".db", ".sqlite3")
	s.names = append(s.names, name)
	return name, nil
}

func (s *stubSnapshotStore) SaveZipFromReader(r io.Reader, filename string) ([]string, error) {
	names := []string{"a.sqlite3", "b.sqlite3"}
	s.names = append(s.names, names...)
	return names, nil
}

func (s *stubSnapshotStore) CloseAndRestore(name, dest string) error {
	if _, err := s.Path(name); err != nil {
		return err
	}
	s.meta.LastLoaded = name
	return nil
}

func TestBackupServiceCreate(t *testing.T) {
	store := &stubSnapshotStore{dir: t.TempDir()}
	svc := NewBackupService(store, "/data/divoene.sqlite3", nil)

	info, err := svc.Create(context.Background())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if info.Name != "divoene-20060102-150405.sqlite3" {
		t.Errorf("unexpected name %q", info.Name)
	}
}

func TestBackupServiceListAndMeta(t *testing.T) {
	store := &stubSnapshotStore{meta: domain.BackupMeta{LastLoaded: "divoene-20060102-150405.sqlite3"}}
	store.names = []string{"divoene-20060102-150405.sqlite3", "divoene-20060101-100000.sqlite3"}
	svc := NewBackupService(store, "/data/divoene.sqlite3", nil)

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	if !list[0].LastLoaded {
		t.Error("last loaded flag not applied")
	}
	if got := svc.Meta(context.Background()); got.LastLoaded != "divoene-20060102-150405.sqlite3" {
		t.Errorf("meta last loaded mismatch: %q", got.LastLoaded)
	}
}

func TestBackupServiceRestoreTriggersRestart(t *testing.T) {
	store := &stubSnapshotStore{names: []string{"x.sqlite3"}}
	restarted := false
	svc := NewBackupService(store, "/data/divoene.sqlite3", func() { restarted = true })

	if err := svc.RestoreLive(context.Background(), "x.sqlite3"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !restarted {
		t.Error("restart callback not invoked")
	}
	if store.meta.LastLoaded != "x.sqlite3" {
		t.Error("snapshot not marked as loaded")
	}
}

func TestBackupServiceRestoreUnknown(t *testing.T) {
	store := &stubSnapshotStore{}
	svc := NewBackupService(store, "/data/divoene.sqlite3", nil)
	if err := svc.RestoreLive(context.Background(), "missing.sqlite3"); err == nil {
		t.Fatal("expected error for unknown snapshot")
	}
}
