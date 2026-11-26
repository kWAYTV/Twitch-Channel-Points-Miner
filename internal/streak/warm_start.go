package streak

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kWAYTV/twitch-points-miner/internal/domain"
	"github.com/kWAYTV/twitch-points-miner/internal/jsonfile"
)

const warmStartCacheVersion = 1

type warmStartFile struct {
	Version int              `json:"version"`
	Entries []WarmStartEntry `json:"entries"`
}

type WarmStartCache struct {
	accountName string
	path        string
	mu          sync.Mutex
	entries     map[string]WarmStartEntry
	dirty       bool
}

func NewWarmStartCache(accountName, path string) *WarmStartCache {
	return &WarmStartCache{
		accountName: strings.ToLower(strings.TrimSpace(accountName)),
		path:        path,
		entries:     make(map[string]WarmStartEntry),
	}
}

func LoadWarmStartCache(path, accountName string) *WarmStartCache {
	cache := NewWarmStartCache(accountName, path)
	if path == "" {
		return cache
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return cache
	}

	var payload warmStartFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return cache
	}
	if payload.Version != 0 && payload.Version != warmStartCacheVersion {
		return cache
	}

	cache.entries = entriesForAccount(payload.Entries, accountName)
	return cache
}

func (c *WarmStartCache) key(streamerLogin string) string {
	return strings.ToLower(strings.TrimSpace(streamerLogin))
}

func (c *WarmStartCache) ResolvedEntryFor(streamer *domain.Streamer, now time.Time) (WarmStartEntry, bool) {
	if c == nil || streamer == nil {
		return WarmStartEntry{}, false
	}

	c.mu.Lock()
	entry, ok := c.entries[c.key(streamer.Username)]
	c.mu.Unlock()
	if !ok || !entryResolved(entry, streamer, now) {
		return WarmStartEntry{}, false
	}
	return entry, true
}

func (c *WarmStartCache) Get(streamerLogin string) (WarmStartEntry, bool) {
	if c == nil {
		return WarmStartEntry{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[c.key(streamerLogin)]
	return entry, ok
}

func (c *WarmStartCache) UpdateFromStreamer(streamer *domain.Streamer, checkedAt time.Time) {
	if c == nil || streamer == nil {
		return
	}
	if checkedAt.IsZero() {
		checkedAt = time.Now()
	}

	entry := entryFromStreamer(c.accountName, streamer, checkedAt)

	key := c.key(streamer.Username)
	c.mu.Lock()
	current, exists := c.entries[key]
	if !exists || current != entry {
		c.entries[key] = entry
		c.dirty = true
	}
	c.mu.Unlock()
}

func (c *WarmStartCache) SaveIfDirty() error {
	if c == nil || c.path == "" {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.dirty {
		return nil
	}
	entries := make([]WarmStartEntry, 0, len(c.entries))
	for _, entry := range c.entries {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StreamerLogin < entries[j].StreamerLogin
	})

	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	if err := jsonfile.SaveJSON(c.path, warmStartFile{
		Version: warmStartCacheVersion,
		Entries: entries,
	}); err != nil {
		return err
	}

	c.dirty = false
	return nil
}
