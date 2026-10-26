// Package vault reads and writes the encrypted DataVault file.
package vault

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCorrupt            = errors.New("vault file is corrupt or not a DataVault file")
	ErrExists             = errors.New("vault already exists")
	ErrLocked             = errors.New("vault is locked")
	ErrNotFound           = errors.New("item not found")
	ErrInvalidItem        = errors.New("invalid item")
)

type payload struct {
	Items []Item `json:"items"`
	// Notes are the Unspoken entries.
	Notes []Note `json:"notes,omitempty"`
	// Tasks are the Siege boards, one per day.
	Tasks []Task `json:"tasks,omitempty"`
	// Files are the crypt's documents; their contents live in the crypt folder.
	Files []File `json:"files,omitempty"`
	// SealPIN is the Argon2id hash of the PIN that opens sealed items.
	SealPIN string `json:"seal_pin,omitempty"`
}

// Vault is an unlocked vault. Its DEK lives only in memory until Lock.
type Vault struct {
	mu      sync.Mutex
	path    string
	header  []byte
	wrapped []byte
	dek     []byte
	data    payload
}

// Exists reports whether a vault file is present at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Create writes a new, empty vault protected by keyword and password.
func Create(path string, keyword, password []byte, p KDFParams) (*Vault, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err == nil {
		return nil, ErrExists
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	salt := randomBytes(saltSize)
	header := encodeHeader(p, salt)
	dek := randomBytes(keySize)

	kek := deriveKEK(keyword, password, salt, p)
	defer wipe(kek)
	wrapped, err := seal(kek, dek, header)
	if err != nil {
		return nil, err
	}

	v := &Vault{path: path, header: header, wrapped: wrapped, dek: dek, data: payload{Items: []Item{}}}
	if err := v.save(v.data); err != nil {
		v.Lock()
		return nil, err
	}
	return v, nil
}

// Open decrypts the vault at path. Wrong credentials return ErrInvalidCredentials
// without saying which of the two was wrong.
func Open(path string, keyword, password []byte) (*Vault, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) < headerSize+wrappedSize {
		return nil, ErrCorrupt
	}
	header := raw[:headerSize]
	wrapped := raw[headerSize : headerSize+wrappedSize]
	box := raw[headerSize+wrappedSize:]

	p, salt, err := decodeHeader(header)
	if err != nil {
		return nil, err
	}
	kek := deriveKEK(keyword, password, salt, p)
	defer wipe(kek)
	dek, err := open(kek, wrapped, header)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	plain, err := open(dek, box, header)
	if err != nil {
		wipe(dek)
		return nil, ErrCorrupt
	}
	defer wipe(plain)

	var data payload
	if err := json.Unmarshal(plain, &data); err != nil {
		wipe(dek)
		return nil, ErrCorrupt
	}
	if data.Items == nil {
		data.Items = []Item{}
	}
	return &Vault{
		path:    path,
		header:  slices.Clone(header),
		wrapped: slices.Clone(wrapped),
		dek:     dek,
		data:    data,
	}, nil
}

// Lock wipes the DEK and drops the decrypted items. The Vault is unusable after.
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()
	wipe(v.dek)
	v.dek = nil
	v.data = payload{}
}

// List returns item summaries matching query (name or tag, case-insensitive),
// sorted by name.
func (v *Vault) List(query string) ([]Summary, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return nil, ErrLocked
	}
	q := strings.ToLower(strings.TrimSpace(query))
	out := []Summary{}
	for _, it := range v.data.Items {
		if q == "" || strings.Contains(strings.ToLower(it.Name), q) || slices.ContainsFunc(it.Tags, func(t string) bool {
			return strings.Contains(t, q)
		}) {
			out = append(out, it.summary())
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}

func (v *Vault) Get(id string) (Item, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Item{}, ErrLocked
	}
	i := v.index(id)
	if i < 0 {
		return Item{}, ErrNotFound
	}
	return v.data.Items[i], nil
}

// Add stores a new item and returns it with its ID and timestamps set.
func (v *Vault) Add(it Item) (Item, error) {
	if err := it.normalize(); err != nil {
		return Item{}, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Item{}, ErrLocked
	}
	now := time.Now().UTC()
	it.ID = hex.EncodeToString(randomBytes(8))
	it.CreatedAt, it.UpdatedAt = now, now

	next := v.data
	next.Items = append(slices.Clone(v.data.Items), it)
	if err := v.save(next); err != nil {
		return Item{}, err
	}
	v.data = next
	return it, nil
}

// Update replaces the item with the given id, keeping its creation time.
func (v *Vault) Update(id string, it Item) (Item, error) {
	if err := it.normalize(); err != nil {
		return Item{}, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Item{}, ErrLocked
	}
	i := v.index(id)
	if i < 0 {
		return Item{}, ErrNotFound
	}
	it.ID = id
	it.CreatedAt = v.data.Items[i].CreatedAt
	// only Renewed moves the last renewal
	it.RenewedAt = v.data.Items[i].RenewedAt
	it.UpdatedAt = time.Now().UTC()

	next := v.data
	next.Items = slices.Clone(v.data.Items)
	next.Items[i] = it
	if err := v.save(next); err != nil {
		return Item{}, err
	}
	v.data = next
	return it, nil
}

func (v *Vault) Delete(id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return ErrLocked
	}
	i := v.index(id)
	if i < 0 {
		return ErrNotFound
	}
	next := v.data
	next.Items = slices.Delete(slices.Clone(v.data.Items), i, i+1)
	if err := v.save(next); err != nil {
		return err
	}
	v.data = next
	return nil
}

// SealPIN returns the stored seal PIN hash, "" when none is set.
func (v *Vault) SealPIN() (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return "", ErrLocked
	}
	return v.data.SealPIN, nil
}

// SetSealPIN stores a new seal PIN hash. The caller hashes and checks the PIN.
func (v *Vault) SetSealPIN(hash string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return ErrLocked
	}
	next := v.data
	next.SealPIN = hash
	if err := v.save(next); err != nil {
		return err
	}
	v.data = next
	return nil
}

func (v *Vault) index(id string) int {
	return slices.IndexFunc(v.data.Items, func(it Item) bool { return it.ID == id })
}

// save encrypts data and replaces the file atomically. The caller holds v.mu
// (or owns v exclusively) and commits data to v only if save succeeds.
func (v *Vault) save(data payload) error {
	plain, err := json.Marshal(data)
	if err != nil {
		return err
	}
	defer wipe(plain)
	box, err := seal(v.dek, plain, v.header)
	if err != nil {
		return err
	}
	file := make([]byte, 0, len(v.header)+len(v.wrapped)+len(box))
	file = append(file, v.header...)
	file = append(file, v.wrapped...)
	file = append(file, box...)
	return writeAtomic(v.path, file)
}

// writeAtomic writes data next to path and renames it into place, keeping the
// previous version as path+".bak".
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".vault-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if prev, err := os.ReadFile(path); err == nil {
		if err := os.WriteFile(path+".bak", prev, 0o600); err != nil {
			return err
		}
	}
	return os.Rename(tmp.Name(), path)
}
