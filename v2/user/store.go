package user

import (
	json "encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/0187773933/MastersCloset/v2/config"
	"github.com/0187773933/MastersCloset/v2/encryption"
	bleve "github.com/blevesearch/bleve/v2"
	bolt "github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"
)

// Bucket names used by the user store.
const (
	bucketUsers        = "users"
	bucketUsernames    = "usernames"
	bucketBarcodes     = "barcodes"
	bucketUpload       = "remote-upload" // drainable pool consumed by remotesync
	bucketMisc         = "misc"          // counters (e.g. virtual-barcode-index)
	bucketCheckInIndex = "checkin-index" // check-in ULID -> user UUID
)

// SearchItem is what we index into Bleve for name search.
type SearchItem struct {
	UUID string
	Name string
}

// SaveOptions tunes a Save. Remote mirrors the record into the remote-upload
// pool (v1 had a whole duplicate SaveLocal method just to skip this). Reindex
// updates the Bleve search index (default on; set SkipReindex to bypass).
type SaveOptions struct {
	Remote      bool
	SkipReindex bool
}

// Store is the persistence boundary for users. It reads config (notably the
// encryption key and timezone) live from the Manager, so config edits are
// observed without reconstructing the store.
type Store struct {
	db    *bolt.DB
	cfg   *config.Manager
	index bleve.Index // may be nil if search is disabled
}

// NewStore wires a store to its db, config manager, and (optional) search index.
func NewStore(db *bolt.DB, cfg *config.Manager, index bleve.Index) *Store {
	db.Update(func(tx *bolt.Tx) error {
		for _, b := range []string{bucketUsers, bucketUsernames, bucketBarcodes, bucketUpload, bucketMisc, bucketCheckInIndex} {
			tx.CreateBucketIfNotExists([]byte(b))
		}
		return nil
	})
	return &Store{db: db, cfg: cfg, index: index}
}

func (s *Store) encryptionKey() string { return s.cfg.Snapshot().BoltDBEncryptionKey }

// location resolves the configured timezone, falling back to local.
func (s *Store) location() *time.Location {
	if tz := s.cfg.Snapshot().TimeZone; tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	return time.Local
}

func (s *Store) now() time.Time { return time.Now().In(s.location()) }

// dateString / timeString match v1's on-disk formats so data is interchangeable.
func dateString(t time.Time) string { return strings.ToUpper(t.Format("02Jan2006")) }
func timeString(t time.Time) string { return t.Format("15:04:05.000") }

// decode decrypts and unmarshals a stored user record.
func (s *Store) decode(raw []byte) (User, bool) {
	var u User
	if raw == nil {
		return u, false
	}
	plain := encryption.ChaChaDecryptBytes(s.encryptionKey(), raw)
	if plain == nil {
		return u, false
	}
	if err := json.Unmarshal(plain, &u); err != nil {
		return u, false
	}
	return u, true
}

// Create makes a new user with the given (raw) username and persists it.
func (s *Store) Create(username string) (User, error) {
	now := s.now()
	u := User{
		UUID:         uuid.NewV4().String(),
		Username:     username,
		Verified:     false,
		FamilySize:   1,
		CreatedDate:  dateString(now),
		CreatedTime:  timeString(now),
		ModifiedDate: dateString(now),
		ModifiedTime: timeString(now),
	}
	if err := s.Save(&u, SaveOptions{Remote: true}); err != nil {
		return u, err
	}
	return u, nil
}

// Get returns a user by UUID.
func (s *Store) Get(userUUID string) (User, bool) {
	var u User
	var ok bool
	s.db.View(func(tx *bolt.Tx) error {
		u, ok = s.decode(tx.Bucket([]byte(bucketUsers)).Get([]byte(userUUID)))
		return nil
	})
	return u, ok
}

// GetByUsername resolves the username index then loads the user.
func (s *Store) GetByUsername(username string) (User, bool) {
	var u User
	var ok bool
	s.db.View(func(tx *bolt.Tx) error {
		id := tx.Bucket([]byte(bucketUsernames)).Get([]byte(username))
		if id == nil {
			return nil
		}
		u, ok = s.decode(tx.Bucket([]byte(bucketUsers)).Get(id))
		return nil
	})
	return u, ok
}

// GetByBarcode resolves the barcode index then loads the user.
func (s *Store) GetByBarcode(barcode string) (User, bool) {
	var u User
	var ok bool
	s.db.View(func(tx *bolt.Tx) error {
		id := tx.Bucket([]byte(bucketBarcodes)).Get([]byte(barcode))
		if id == nil {
			return nil
		}
		u, ok = s.decode(tx.Bucket([]byte(bucketUsers)).Get(id))
		return nil
	})
	return u, ok
}

// UsernameExists reports whether a username is taken, and its UUID if so.
func (s *Store) UsernameExists(username string) (bool, string) {
	var id string
	s.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket([]byte(bucketUsernames)).Get([]byte(username)); v != nil {
			id = string(v)
		}
		return nil
	})
	return id != "", id
}

// Save persists a user: maintains the username & barcode indexes, fixes the
// modified timestamps (v1 wrongly overwrote CreatedDate on every save), and
// updates the search index. This single method replaces v1's Save + SaveLocal.
func (s *Store) Save(u *User, opts SaveOptions) error {
	now := s.now()
	FormatUsername(u)

	// Locals (matching the configured zipcode) are auto-verified.
	if zip := strings.TrimSpace(u.Identity.Address.ZipCode); zip != "" {
		target := s.cfg.Snapshot().AutoVerifyZipcode
		if target == "" {
			target = "45424"
		}
		if zip == target {
			u.Verified = true
		}
	}
	if u.CreatedDate == "" {
		u.CreatedDate = dateString(now)
		u.CreatedTime = timeString(now)
	}
	u.ModifiedDate = dateString(now)
	u.ModifiedTime = timeString(now)

	key := s.encryptionKey()
	payload, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}
	enc := encryption.ChaChaEncryptBytes(key, payload)

	err = s.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte(bucketUsers))
		usernames := tx.Bucket([]byte(bucketUsernames))
		barcodes := tx.Bucket([]byte(bucketBarcodes))

		// Reconcile against the previously stored record: drop a renamed
		// username index entry, and delete barcodes the admin removed so they
		// stop resolving to this user.
		if existing, ok := s.decode(users.Get([]byte(u.UUID))); ok {
			if existing.Username != u.Username {
				usernames.Delete([]byte(existing.Username))
			}
			kept := map[string]bool{}
			for _, bc := range u.Barcodes {
				kept[bc] = true
			}
			for _, bc := range existing.Barcodes {
				if bc != "" && !kept[bc] {
					barcodes.Delete([]byte(bc))
				}
			}
		}
		usernames.Put([]byte(u.Username), []byte(u.UUID))
		for _, bc := range u.Barcodes {
			if bc != "" {
				barcodes.Put([]byte(bc), []byte(u.UUID))
			}
		}
		users.Put([]byte(u.UUID), enc)
		if opts.Remote {
			tx.Bucket([]byte(bucketUpload)).Put([]byte(u.UUID), enc)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if !opts.SkipReindex && s.index != nil {
		s.index.Index(u.UUID, SearchItem{UUID: u.UUID, Name: u.SearchString})
	}
	return nil
}

// Delete removes a user and every index entry pointing at it. v1's Delete was
// entirely commented out — a silent no-op.
func (s *Store) Delete(userUUID string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte(bucketUsers))
		existing, ok := s.decode(users.Get([]byte(userUUID)))
		if !ok {
			return nil
		}
		tx.Bucket([]byte(bucketUsernames)).Delete([]byte(existing.Username))
		barcodes := tx.Bucket([]byte(bucketBarcodes))
		for _, bc := range existing.Barcodes {
			barcodes.Delete([]byte(bc))
		}
		users.Delete([]byte(userUUID))
		tx.Bucket([]byte(bucketUpload)).Delete([]byte(userUUID))
		return nil
	})
	if err != nil {
		return err
	}
	if s.index != nil {
		s.index.Delete(userUUID)
	}
	return nil
}

// ForEach iterates every decoded user. Used by similarity scans and reports.
func (s *Store) ForEach(fn func(User)) {
	s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketUsers)).ForEach(func(_, v []byte) error {
			if u, ok := s.decode(v); ok {
				fn(u)
			}
			return nil
		})
	})
}
