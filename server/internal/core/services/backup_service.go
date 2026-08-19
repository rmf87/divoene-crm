package services

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/rmf87/divoene/internal/core/domain"
)

// BackupService orchestrates local database snapshots. It owns no
// infrastructure: persistence goes through domain.SnapshotStore and the
// post-restore restart through the injected restart callback.
type BackupService struct {
	store   domain.SnapshotStore
	dbPath  string
	restart func()
}

// NewBackupService wires a SnapshotStore, the live DB path and the process
// restart callback used after a live restore.
func NewBackupService(store domain.SnapshotStore, dbPath string, restart func()) *BackupService {
	return &BackupService{store: store, dbPath: dbPath, restart: restart}
}

// Create makes a new snapshot and returns its file info.
func (s *BackupService) Create(ctx context.Context) (domain.SnapshotInfo, error) {
	path, err := s.store.Create()
	if err != nil {
		return domain.SnapshotInfo{}, err
	}
	return s.infoFor(path)
}

// List returns all snapshots, newest first.
func (s *BackupService) List(ctx context.Context) ([]domain.SnapshotInfo, error) {
	return s.store.List()
}

// Meta returns the persisted history and last-loaded marker.
func (s *BackupService) Meta(ctx context.Context) domain.BackupMeta {
	return s.store.Meta()
}

// Path resolves a snapshot name to its full path.
func (s *BackupService) Path(ctx context.Context, name string) (string, error) {
	return s.store.Path(name)
}

// Delete removes a snapshot.
func (s *BackupService) Delete(ctx context.Context, name string) error {
	return s.store.Delete(name)
}

// SaveFromReader stores a single SQLite stream as a snapshot.
func (s *BackupService) SaveFromReader(ctx context.Context, r io.Reader, filename string) (domain.SnapshotInfo, error) {
	path, err := s.store.SaveFromReader(r, filename)
	if err != nil {
		return domain.SnapshotInfo{}, err
	}
	return s.infoFor(path)
}

// SaveZipFromReader extracts SQLite snapshots from a ZIP stream.
func (s *BackupService) SaveZipFromReader(ctx context.Context, r io.Reader, filename string) ([]domain.SnapshotInfo, error) {
	names, err := s.store.SaveZipFromReader(r, filename)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SnapshotInfo, 0, len(names))
	for _, n := range names {
		info, err := s.infoFor(n)
		if err != nil {
			continue
		}
		out = append(out, info)
	}
	return out, nil
}

// RestoreLive closes the database, restores the snapshot and triggers the
// process restart so the server comes back with the restored data.
func (s *BackupService) RestoreLive(ctx context.Context, name string) error {
	if err := s.store.CloseAndRestore(name, s.dbPath); err != nil {
		return err
	}
	if s.restart != nil {
		s.restart()
	}
	return nil
}

func (s *BackupService) infoFor(path string) (domain.SnapshotInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return domain.SnapshotInfo{}, err
	}
	return domain.SnapshotInfo{
		Name:    filepath.Base(path),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}
