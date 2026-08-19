package domain

import (
	"io"
	"time"
)

// SnapshotInfo represents a local database snapshot exposed by the API.
type SnapshotInfo struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"modified_at"`
	LastLoaded bool      `json:"last_loaded,omitempty"`
}

// BackupHistoryEntry records one backup lifecycle event.
type BackupHistoryEntry struct {
	Name   string    `json:"name"`
	Action string    `json:"action"`
	At     time.Time `json:"at"`
}

// BackupMeta is the persisted snapshot history and "last loaded" marker.
type BackupMeta struct {
	LastLoaded string               `json:"last_loaded"`
	LoadedAt   time.Time            `json:"loaded_at,omitempty"`
	History    []BackupHistoryEntry `json:"history"`
}

// SnapshotStore is the port for local database snapshots (VACUUM INTO / restore).
// Implementations live in internal/infra/storage.
type SnapshotStore interface {
	// Create snapshots the current database and returns the stored path.
	Create() (string, error)
	// List returns all snapshots, newest first.
	List() ([]SnapshotInfo, error)
	// Meta returns the persisted history and last-loaded marker.
	Meta() BackupMeta
	// Path resolves a snapshot name to its full path.
	Path(name string) (string, error)
	// Delete removes a snapshot.
	Delete(name string) error
	// SaveFromReader stores a single SQLite stream as a snapshot.
	SaveFromReader(r io.Reader, filename string) (string, error)
	// SaveZipFromReader extracts SQLite snapshots from a ZIP stream.
	SaveZipFromReader(r io.Reader, filename string) ([]string, error)
	// CloseAndRestore closes the live database, restores the snapshot to dest,
	// clears WAL/SHM sidecar files and marks the snapshot as loaded.
	CloseAndRestore(name, dest string) error
}
