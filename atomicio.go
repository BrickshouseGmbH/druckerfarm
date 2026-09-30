package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// ─── ATOMIC FILE WRITES + ROLLING BACKUP ──────────────────────────────────────
//
// All persistent state (config.json, runtime.json, upload_log.json, sync.json)
// used to be written with a plain os.WriteFile. If the process was killed or the
// disk hiccuped mid-write, the file could be left truncated or empty — losing the
// user's whole configuration. atomicWrite avoids that: it writes to a temporary
// file in the same directory and then renames it over the target. os.Rename is
// atomic on a single volume, so a reader always sees either the complete old file
// or the complete new one, never a half-written mix.
//
// For config.json we additionally keep a small rolling set of timestamped
// backups, so an accidental bad write (or a manual edit gone wrong) can be
// recovered by hand.

// atomicWriteMu serialises writes to the same target across goroutines so two
// concurrent savers cannot interleave their temp files onto one path.
var atomicWriteMu sync.Mutex

// atomicWrite writes data to path via a temp file + rename. It returns any error
// so callers can log it; existing call sites ignore the error just as the old
// os.WriteFile calls did, but the write itself is now crash-safe.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	if path == "" {
		return fmt.Errorf("atomicWrite: empty path")
	}
	atomicWriteMu.Lock()
	defer atomicWriteMu.Unlock()

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		// Fallback: the directory may be read-only for temp creation; fall back to
		// a direct write so we degrade to the old behaviour instead of losing data.
		return os.WriteFile(path, data, perm)
	}
	tmpName := tmp.Name()
	// On any failure past this point, remove the leftover temp file.
	cleanup := func() { tmp.Close(); os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	// Flush to disk before the rename so a crash right after the rename cannot
	// leave an empty-but-named file.
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	_ = os.Chmod(tmpName, perm)
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// backupConfig copies the current config.json to config.json.bak-YYYYmmdd-HHMMSS
// before it is overwritten, keeping only the newest keepConfigBackups files. It
// is a best-effort helper: any error is silently ignored so a backup problem can
// never block the real save.
const keepConfigBackups = 8

func backupConfig() {
	if dataFile == "" {
		return
	}
	cur, err := os.ReadFile(dataFile)
	if err != nil || len(cur) == 0 {
		return // nothing (or nothing useful) to back up yet
	}
	stamp := time.Now().Format("20060102-150405")
	bak := dataFile + ".bak-" + stamp
	_ = os.WriteFile(bak, cur, 0644)
	pruneConfigBackups()
}

// pruneConfigBackups removes the oldest config.json.bak-* files beyond the limit.
func pruneConfigBackups() {
	if dataFile == "" {
		return
	}
	matches, err := filepath.Glob(dataFile + ".bak-*")
	if err != nil || len(matches) <= keepConfigBackups {
		return
	}
	sort.Strings(matches) // timestamp names sort chronologically
	for _, old := range matches[:len(matches)-keepConfigBackups] {
		os.Remove(old)
	}
}
