package cache // or wherever you place it

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chubin/wttr.in/internal/domain"
)

// DiskCacher is a filesystem-backed implementation of Cacher.
// Completed entries are persisted to disk as JSON files.
// In-progress state is kept in-memory (per-process) with notification channels.
type DiskCacher struct {
	dir        string
	mu         sync.RWMutex
	inProgress map[string]*progressInfo
}

type progressInfo struct {
	done  chan struct{}
	entry *domain.CacheEntry // set on completion, nil on removal
}

// NewDiskCacher creates a new disk-backed cache.
// dir will be created if it doesn't exist.
func NewDiskCacher(dir string) (*DiskCacher, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}
	return &DiskCacher{
		dir:        dir,
		inProgress: make(map[string]*progressInfo),
	}, nil
}

func (d *DiskCacher) filePath(key string) string {
	h := sha256.Sum256([]byte(key))
	return filepath.Join(d.dir, hex.EncodeToString(h[:])+".json")
}

// writeToDisk atomically writes the entry using a temp file.
func (d *DiskCacher) writeToDisk(key string, entry *domain.CacheEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	path := d.filePath(key)
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // cleanup
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func (d *DiskCacher) readFromDisk(key string) (*domain.CacheEntry, error) {
	path := d.filePath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	var entry domain.CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal cache entry: %w", err)
	}
	return &entry, nil
}

// Get returns a valid (non-expired) cache entry.
func (d *DiskCacher) Get(key string) *domain.CacheEntry {
	entry, err := d.readFromDisk(key)
	if err != nil {
		return nil
	}
	if time.Now().After(entry.Expires) { // expired
		// optional: lazy delete
		_ = os.Remove(d.filePath(key))
		return nil
	}
	return entry
}

// Set stores the completed entry and notifies waiters.
func (d *DiskCacher) Set(key string, entry domain.CacheEntry) {
	// Persist to disk
	if err := d.writeToDisk(key, &entry); err != nil {
		// In production you might want to log this, but don't fail the Set
		// (the in-memory notification still happens)
	}

	d.mu.Lock()
	if p, ok := d.inProgress[key]; ok {
		entryCopy := entry // defensive copy of value
		p.entry = &entryCopy
		close(p.done)
		delete(d.inProgress, key)
	}
	d.mu.Unlock()
}

// SetInProgressIfNotExists returns true if it successfully marked the key as in-progress.
func (d *DiskCacher) SetInProgressIfNotExists(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.inProgress[key]; ok {
		return false
	}

	d.inProgress[key] = &progressInfo{
		done: make(chan struct{}),
	}
	return true
}

func (d *DiskCacher) IsInProgress(key string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.inProgress[key]
	return ok
}

// WaitForCompletion blocks until the key is no longer in-progress or timeout occurs.
func (d *DiskCacher) WaitForCompletion(key string, maxWait time.Duration) (*domain.CacheEntry, error) {
	d.mu.RLock()
	p, ok := d.inProgress[key]
	d.mu.RUnlock()

	if !ok {
		// Already completed or never started → try disk
		return d.Get(key), nil
	}

	timer := time.NewTimer(maxWait)
	defer timer.Stop()

	select {
	case <-p.done:
		if p.entry != nil {
			return p.entry, nil
		}
		return nil, nil // was removed (e.g. error response)
	case <-timer.C:
		return nil, fmt.Errorf("timeout waiting for cache key %s", key)
	}
}

// Remove deletes the on-disk entry and clears any in-progress state.
func (d *DiskCacher) Remove(key string) {
	_ = os.Remove(d.filePath(key))

	d.mu.Lock()
	if p, ok := d.inProgress[key]; ok {
		p.entry = nil
		close(p.done)
		delete(d.inProgress, key)
	}
	d.mu.Unlock()
}

func (d *DiskCacher) Close() error {
	// No persistent resources (e.g. no DB connection). In-memory state is dropped.
	// You could optionally delete all files here if desired.
	return nil
}
