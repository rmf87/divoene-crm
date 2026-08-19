package storage

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"

	"github.com/rmf87/divoene/internal/core/domain"
)

// SnapshotStore implements domain.SnapshotStore on top of local snapshots.
type SnapshotStore struct {
	db     *sql.DB
	dbPath string
}

// NewSnapshotStore creates a SnapshotStore bound to the live database handle
// and its on-disk path (used for restore).
func NewSnapshotStore(db *sql.DB, dbPath string) *SnapshotStore {
	return &SnapshotStore{db: db, dbPath: dbPath}
}

// Create snapshots the current database via VACUUM INTO.
func (s *SnapshotStore) Create() (string, error) {
	return CreateSnapshot(s.db)
}

// List returns all snapshots, newest first.
func (s *SnapshotStore) List() ([]domain.SnapshotInfo, error) {
	files, err := ListSnapshots()
	if err != nil {
		return nil, err
	}
	out := make([]domain.SnapshotInfo, 0, len(files))
	for _, f := range files {
		out = append(out, domain.SnapshotInfo{
			Name:       f.Name,
			Size:       f.Size,
			ModTime:    f.ModTime,
			LastLoaded: f.LastLoaded,
		})
	}
	return out, nil
}

// Meta returns the persisted history and last-loaded marker.
func (s *SnapshotStore) Meta() domain.BackupMeta {
	meta := LoadMeta()
	return domain.BackupMeta{
		LastLoaded: meta.LastLoaded,
		LoadedAt:   meta.LoadedAt,
		History:    mapHistory(meta.History),
	}
}

func mapHistory(in []BackupHistoryEntry) []domain.BackupHistoryEntry {
	out := make([]domain.BackupHistoryEntry, 0, len(in))
	for _, h := range in {
		out = append(out, domain.BackupHistoryEntry{
			Name:   h.Name,
			Action: h.Action,
			At:     h.At,
		})
	}
	return out
}

// Path resolves a snapshot name to its full path.
func (s *SnapshotStore) Path(name string) (string, error) {
	return SnapshotPath(name)
}

// Delete removes a snapshot.
func (s *SnapshotStore) Delete(name string) error {
	return DeleteSnapshot(name)
}

// SaveFromReader stores a single SQLite stream as a snapshot.
func (s *SnapshotStore) SaveFromReader(r io.Reader, filename string) (string, error) {
	return SaveUploadedSnapshot(r, filename)
}

// SaveZipFromReader extracts SQLite snapshots from a ZIP stream.
func (s *SnapshotStore) SaveZipFromReader(r io.Reader, filename string) ([]string, error) {
	return SaveUploadedZip(r, filename)
}

// CloseAndRestore closes the live database, restores the snapshot to the
// configured path, clears WAL/SHM sidecars and marks the snapshot as loaded.
func (s *SnapshotStore) CloseAndRestore(name string, _ string) error {
	if _, err := SnapshotPath(name); err != nil {
		return err
	}

	if err := s.db.Close(); err != nil {
		return err
	}

	if err := RestoreSnapshot(name, s.dbPath); err != nil {
		return err
	}

	os.Remove(filepath.Clean(s.dbPath) + "-wal")
	os.Remove(filepath.Clean(s.dbPath) + "-shm")

	return MarkLoaded(name)
}
