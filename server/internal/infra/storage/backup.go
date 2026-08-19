package storage

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// defaultBackupDir is where local snapshots are stored unless BACKUP_DIR
// overrides it (used by tests and deployments with custom volumes).
const defaultBackupDir = "/data/backups"

// Dir returns the directory where local snapshots are stored.
func Dir() string {
	if d := os.Getenv("BACKUP_DIR"); d != "" {
		return d
	}
	return defaultBackupDir
}

// metaFileName holds the snapshot history and "last loaded" marker. It lives
// next to the backups so it survives a database restore.
const metaFileName = "backups.json"

// maxZipEntries caps the number of files extracted from a single uploaded ZIP.
const maxZipEntries = 200

// CreateSnapshot creates a VACUUM INTO snapshot of the database
// to the local backup directory and returns the file path.
func CreateSnapshot(db *sql.DB) (string, error) {
	if err := os.MkdirAll(Dir(), 0755); err != nil {
		return "", fmt.Errorf("backup: create dir: %w", err)
	}

	filename := fmt.Sprintf("divoene-%s.sqlite3", time.Now().Format("20060102-150405"))
	snapshotPath := filepath.Join(Dir(), filename)

	if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", snapshotPath)); err != nil {
		return "", fmt.Errorf("backup: vacuum: %w", err)
	}

	RecordAction(filename, "create")
	return snapshotPath, nil
}

// ListSnapshots returns all backup files in the backup directory, newest first.
// The last-loaded marker from the metadata file is applied to the matching file.
func ListSnapshots() ([]BackupFile, error) {
	entries, err := os.ReadDir(Dir())
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupFile{}, nil
		}
		return nil, fmt.Errorf("backup: read dir: %w", err)
	}

	meta := LoadMeta()
	var files []BackupFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == metaFileName || strings.HasPrefix(name, ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, BackupFile{
			Name:       name,
			Size:       info.Size(),
			ModTime:    info.ModTime(),
			LastLoaded: name == meta.LastLoaded,
		})
	}

	// Sort newest first
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].ModTime.After(files[i].ModTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	return files, nil
}

// SnapshotPath returns the full path for a backup filename.
func SnapshotPath(name string) (string, error) {
	if name == metaFileName || strings.HasPrefix(name, ".") {
		return "", fmt.Errorf("backup: file not found: %s", name)
	}
	path := filepath.Join(Dir(), name)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("backup: file not found: %s", name)
	}
	return path, nil
}

// RestoreSnapshot copies a backup file to the destination path atomically.
func RestoreSnapshot(name, destPath string) error {
	src, err := SnapshotPath(name)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("backup: open snapshot: %w", err)
	}
	defer in.Close()

	tmp, err := os.CreateTemp(filepath.Dir(destPath), ".restore-*")
	if err != nil {
		return fmt.Errorf("backup: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return fmt.Errorf("backup: copy snapshot: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("backup: sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("backup: close temp: %w", err)
	}

	if err := os.Rename(tmpName, destPath); err != nil {
		return fmt.Errorf("backup: replace dest: %w", err)
	}
	return nil
}

// MarkLoaded records a snapshot as the last loaded one.
func MarkLoaded(name string) error {
	if _, err := SnapshotPath(name); err != nil {
		return err
	}
	meta := LoadMeta()
	meta.LastLoaded = name
	meta.LoadedAt = time.Now().UTC()
	meta.History = append(meta.History, BackupHistoryEntry{
		Name:   name,
		Action: "restore",
		At:     time.Now().UTC(),
	})
	return SaveMeta(meta)
}

// DeleteSnapshot removes a backup file from the backup directory.
func DeleteSnapshot(name string) error {
	path, err := SnapshotPath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("backup: delete: %w", err)
	}

	meta := LoadMeta()
	if meta.LastLoaded == name {
		meta.LastLoaded = ""
		meta.LoadedAt = time.Time{}
		_ = SaveMeta(meta)
	}
	return nil
}

// unsafeName matches path separators and other characters that must not
// appear in a user-supplied backup filename.
var unsafeName = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// isSQLiteName reports whether a filename looks like a SQLite database.
func isSQLiteName(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".sqlite3" || ext == ".sqlite" || ext == ".db"
}

// uniqueSnapshotName returns base if it does not collide, otherwise a
// timestamped variant that is guaranteed not to exist.
func uniqueSnapshotName(base string) (string, error) {
	if _, err := os.Stat(filepath.Join(Dir(), base)); os.IsNotExist(err) {
		return base, nil
	}
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("%s-%s%s", stem, time.Now().Format("150405.000000000"), ext)
		if _, err := os.Stat(filepath.Join(Dir(), name)); os.IsNotExist(err) {
			return name, nil
		}
		time.Sleep(time.Millisecond)
	}
	return "", fmt.Errorf("backup: could not allocate unique name for %s", base)
}

// saveSnapshotFile validates the SQLite magic header and writes the stream to
// the backup directory under a unique name.
func saveSnapshotFile(src io.Reader, base string) (string, error) {
	if err := os.MkdirAll(Dir(), 0755); err != nil {
		return "", fmt.Errorf("backup: create dir: %w", err)
	}

	// Read first 16 bytes to validate SQLite magic header
	header := make([]byte, 16)
	n, err := io.ReadFull(src, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("backup: read header: %w", err)
	}
	if n < 16 || string(header[:15]) != "SQLite format 3" {
		return "", fmt.Errorf("backup: not a valid SQLite database")
	}

	name, err := uniqueSnapshotName(base)
	if err != nil {
		return "", err
	}
	destPath := filepath.Join(Dir(), name)

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("backup: create file: %w", err)
	}
	defer out.Close()

	if _, err := out.Write(header[:n]); err != nil {
		return "", fmt.Errorf("backup: write header: %w", err)
	}
	if _, err := io.Copy(out, src); err != nil {
		return "", fmt.Errorf("backup: copy body: %w", err)
	}

	RecordAction(name, "upload")
	return name, nil
}

// SaveUploadedSnapshot reads from src and stores it in the backup directory.
// The original filename is sanitized; the stored name is returned.
func SaveUploadedSnapshot(src io.Reader, originalName string) (string, error) {
	base := filepath.Base(originalName)
	base = unsafeName.ReplaceAllString(base, "_")
	if base == "" || base == "." || base == "_" {
		base = fmt.Sprintf("upload-%s.sqlite3", time.Now().Format("20060102-150405"))
	}
	if !isSQLiteName(base) {
		base += ".sqlite3"
	}
	return saveSnapshotFile(src, base)
}

// SaveUploadedZip extracts all SQLite databases from a ZIP upload and stores
// each one in the backup directory. Returns the stored names.
func SaveUploadedZip(src io.Reader, originalName string) ([]string, error) {
	tmp, err := os.CreateTemp("", "backup-upload-*.zip")
	if err != nil {
		return nil, fmt.Errorf("backup: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	size, err := io.Copy(tmp, src)
	if err != nil {
		tmp.Close()
		return nil, fmt.Errorf("backup: read upload: %w", err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("backup: seek temp: %w", err)
	}

	zr, err := zip.NewReader(tmp, size)
	if err != nil {
		tmp.Close()
		return nil, fmt.Errorf("backup: invalid zip: %w", err)
	}

	var saved []string
	for _, f := range zr.File {
		if len(saved) >= maxZipEntries {
			break
		}
		if f.FileInfo().IsDir() {
			continue
		}
		base := filepath.Base(f.Name)
		if base == "" || base == "." {
			continue
		}
		base = unsafeName.ReplaceAllString(base, "_")
		if !isSQLiteName(base) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		name, err := saveSnapshotFile(rc, base)
		rc.Close()
		if err != nil {
			continue
		}
		saved = append(saved, name)
	}
	tmp.Close()

	if len(saved) == 0 {
		return nil, fmt.Errorf("backup: no valid sqlite files found in zip")
	}
	return saved, nil
}

// BackupHistoryEntry records one backup lifecycle event.
type BackupHistoryEntry struct {
	Name   string    `json:"name"`
	Action string    `json:"action"`
	At     time.Time `json:"at"`
}

// BackupMeta is the persisted snapshot history and "last loaded" marker.
type BackupMeta struct {
	LastLoaded string              `json:"last_loaded"`
	LoadedAt   time.Time           `json:"loaded_at,omitempty"`
	History    []BackupHistoryEntry `json:"history"`
}

func metaPath() string { return filepath.Join(Dir(), metaFileName) }

// LoadMeta reads the metadata file. A missing or corrupt file yields empty meta.
func LoadMeta() BackupMeta {
	data, err := os.ReadFile(metaPath())
	if err != nil {
		return BackupMeta{History: []BackupHistoryEntry{}}
	}
	var meta BackupMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return BackupMeta{History: []BackupHistoryEntry{}}
	}
	if meta.History == nil {
		meta.History = []BackupHistoryEntry{}
	}
	return meta
}

// SaveMeta writes the metadata file atomically.
func SaveMeta(meta BackupMeta) error {
	if err := os.MkdirAll(Dir(), 0755); err != nil {
		return fmt.Errorf("backup: create dir: %w", err)
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("backup: marshal meta: %w", err)
	}
	tmp, err := os.CreateTemp(Dir(), ".backups-*.json")
	if err != nil {
		return fmt.Errorf("backup: meta temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("backup: write meta: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("backup: close meta: %w", err)
	}
	if err := os.Rename(tmpName, metaPath()); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("backup: replace meta: %w", err)
	}
	return nil
}

// RecordAction appends an entry to the backup history.
func RecordAction(name, action string) {
	meta := LoadMeta()
	meta.History = append(meta.History, BackupHistoryEntry{
		Name:   name,
		Action: action,
		At:     time.Now().UTC(),
	})
	_ = SaveMeta(meta)
}

// BackupFile represents a local backup snapshot.
type BackupFile struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"modified_at"`
	LastLoaded bool      `json:"last_loaded,omitempty"`
}
