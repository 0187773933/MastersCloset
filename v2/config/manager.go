package config

import (
	json "encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	bolt "github.com/boltdb/bolt"
)

const (
	bucketName = "config"
	currentKey = "current"
)

// Manager owns the single source of truth for configuration. Every reader takes
// a Snapshot (an RLock'd copy), so a successful Update is observed by all
// subsequent handler calls without a restart — this is the "server monolith var"
// the README asked for.
type Manager struct {
	mu  sync.RWMutex
	cfg Config
	db  *bolt.DB
}

// ParseSeedFile reads a JSON config file used to bootstrap a fresh database.
func ParseSeedFile(path string) (Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("read seed file %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("parse seed file %q: %w", path, err)
	}
	return c, nil
}

// NewManager wraps an already-open BoltDB.
func NewManager(db *bolt.DB) *Manager { return &Manager{db: db} }

// LoadOrSeed loads the persisted config from BoltDB. If none exists yet it seeds
// the db from the provided file-parsed config and returns seeded=true. The
// bootstrap-only fields (db path/key/bleve path, fingerprint) from seed always
// win, since they are how we actually opened the db.
func (m *Manager) LoadOrSeed(seed Config) (seeded bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if seed.AutoVerifyZipcode == "" {
		seed.AutoVerifyZipcode = "45424" // default; editable in the settings panel
	}
	err = m.db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte(bucketName))
		if e != nil {
			return e
		}
		raw := b.Get([]byte(currentKey))
		if raw == nil {
			m.cfg = seed
			seeded = true
			data, _ := json.MarshalIndent(seed, "", "  ")
			return b.Put([]byte(currentKey), data)
		}
		var loaded Config
		if e := json.Unmarshal(raw, &loaded); e != nil {
			return e
		}
		// Bootstrap-only fields come from the seed we actually booted with.
		loaded.BoltDBPath = seed.BoltDBPath
		loaded.BoltDBEncryptionKey = seed.BoltDBEncryptionKey
		loaded.BleveSearchPath = seed.BleveSearchPath
		if seed.FingerPrint != "" {
			loaded.FingerPrint = seed.FingerPrint
		}
		if loaded.AutoVerifyZipcode == "" {
			loaded.AutoVerifyZipcode = seed.AutoVerifyZipcode // backfill for older dbs
		}
		m.cfg = loaded
		return nil
	})
	return
}

// Snapshot returns a copy of the current config, safe to read without locking.
func (m *Manager) Snapshot() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// FlatSnapshot returns the current config as dotted-path -> value, the shape the
// settings panel renders against.
func (m *Manager) FlatSnapshot() map[string]interface{} {
	return toFlat(m.Snapshot())
}

// ApplyEdits coerces and applies a flat map of dotted-path -> raw value onto the
// current config, persists it, and swaps the in-memory snapshot. It returns the
// human labels of any restart-required fields that changed. ReadOnly fields in
// the edit set are ignored.
func (m *Manager) ApplyEdits(edits map[string]interface{}) (restart []string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta := metaByPath()
	flat := toFlat(m.cfg)
	for path, raw := range edits {
		fm, known := meta[path]
		if !known || fm.ReadOnly {
			continue // unknown or bootstrap-only — never editable here
		}
		coerced, cerr := coerce(fm.Kind, raw)
		if cerr != nil {
			return nil, fmt.Errorf("%s: %w", fm.Label, cerr)
		}
		flat[path] = coerced
	}

	next, berr := fromFlat(flat)
	if berr != nil {
		return nil, berr
	}
	// Re-assert bootstrap-only fields no matter what.
	next.BoltDBPath = m.cfg.BoltDBPath
	next.BoltDBEncryptionKey = m.cfg.BoltDBEncryptionKey
	next.BleveSearchPath = m.cfg.BleveSearchPath
	next.FingerPrint = m.cfg.FingerPrint

	if verr := validate(next); verr != nil {
		return nil, verr
	}
	restart = diffRestart(m.cfg, next)

	data, _ := json.MarshalIndent(next, "", "  ")
	if err = m.db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte(bucketName))
		if e != nil {
			return e
		}
		return b.Put([]byte(currentKey), data)
	}); err != nil {
		return nil, err
	}
	m.cfg = next
	return restart, nil
}

// validate enforces the few invariants that would otherwise break the server.
func validate(c Config) error {
	if c.ServerPort == "" {
		return fmt.Errorf("server_port is required")
	}
	if _, err := strconv.Atoi(c.ServerPort); err != nil {
		return fmt.Errorf("server_port must be numeric")
	}
	if c.CheckInCoolOffDays < 0 {
		return fmt.Errorf("check_in_cooloff_days cannot be negative")
	}
	if c.LevenshteinDistanceThreshold < 0 {
		return fmt.Errorf("levenshtein_distance_threshold cannot be negative")
	}
	if c.TimeZone != "" {
		if _, err := time.LoadLocation(c.TimeZone); err != nil {
			return fmt.Errorf("time_zone %q is not a valid IANA zone", c.TimeZone)
		}
	}
	return nil
}

// diffRestart returns labels of restart-required fields that differ.
func diffRestart(oldC, newC Config) []string {
	of, nf := toFlat(oldC), toFlat(newC)
	var changed []string
	for _, f := range Fields() {
		if !f.RestartRequired {
			continue
		}
		if fmt.Sprint(of[f.Path]) != fmt.Sprint(nf[f.Path]) {
			changed = append(changed, f.Label)
		}
	}
	return changed
}

// coerce converts a raw value (often a form string) into the Go type a field
// expects, based on its declared Kind.
func coerce(kind string, raw interface{}) (interface{}, error) {
	switch kind {
	case "int":
		switch v := raw.(type) {
		case float64:
			return int(v), nil
		case int:
			return v, nil
		case string:
			if strings.TrimSpace(v) == "" {
				return 0, nil
			}
			return strconv.Atoi(strings.TrimSpace(v))
		}
		return 0, fmt.Errorf("expected integer")
	case "float":
		switch v := raw.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case string:
			if strings.TrimSpace(v) == "" {
				return 0.0, nil
			}
			return strconv.ParseFloat(strings.TrimSpace(v), 64)
		}
		return 0.0, fmt.Errorf("expected number")
	case "bool":
		switch v := raw.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(strings.TrimSpace(v))
		}
		return false, fmt.Errorf("expected boolean")
	case "list":
		switch v := raw.(type) {
		case []interface{}:
			out := make([]string, 0, len(v))
			for _, e := range v {
				out = append(out, fmt.Sprint(e))
			}
			return out, nil
		case string:
			var out []string
			for _, part := range strings.FieldsFunc(v, func(r rune) bool { return r == '\n' || r == ',' }) {
				if p := strings.TrimSpace(part); p != "" {
					out = append(out, p)
				}
			}
			return out, nil
		}
		return []string{}, nil
	default: // string
		return fmt.Sprint(raw), nil
	}
}

// toFlat marshals a Config to a dotted-path map of scalar values.
func toFlat(c Config) map[string]interface{} {
	data, _ := json.Marshal(c)
	var nested map[string]interface{}
	json.Unmarshal(data, &nested)
	out := map[string]interface{}{}
	flatten("", nested, out)
	return out
}

func flatten(prefix string, in map[string]interface{}, out map[string]interface{}) {
	for k, v := range in {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		if child, ok := v.(map[string]interface{}); ok {
			flatten(path, child, out)
		} else {
			out[path] = v
		}
	}
}

// fromFlat rebuilds a Config from a dotted-path map.
func fromFlat(flat map[string]interface{}) (Config, error) {
	nested := map[string]interface{}{}
	for path, val := range flat {
		setIn(nested, strings.Split(path, "."), val)
	}
	data, _ := json.Marshal(nested)
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("rebuild config: %w", err)
	}
	return c, nil
}

func setIn(m map[string]interface{}, keys []string, val interface{}) {
	if len(keys) == 1 {
		m[keys[0]] = val
		return
	}
	child, ok := m[keys[0]].(map[string]interface{})
	if !ok {
		child = map[string]interface{}{}
		m[keys[0]] = child
	}
	setIn(child, keys[1:], val)
}
