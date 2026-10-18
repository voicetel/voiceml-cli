// Package config loads and persists the CLI's per-user configuration.
//
// The on-disk format is TOML at ~/.voiceml/config.toml. The file is
// written atomically (write to a temp sibling, fsync, rename) so we
// never leave a half-written config behind, and is locked down to 0600.
package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// DefaultBaseURL is the production VoiceML API endpoint. Stored as a
// constant rather than imported from the SDK to keep this package
// dependency-free.
const DefaultBaseURL = "https://voiceml.voicetel.com"

// Config is the persisted shape.
type Config struct {
	AccountSid string `toml:"account_sid,omitempty"`
	APIKey     string `toml:"api_key,omitempty"`
	BaseURL    string `toml:"base_url,omitempty"`
}

// ErrNotFound is returned by Load when the file does not exist.
var ErrNotFound = errors.New("config: file does not exist")

// userHomeDir is a function variable so tests can simulate the
// `os.UserHomeDir` error path (which is otherwise hard to trigger on a
// normal dev/CI machine where HOME is always set).
var userHomeDir = os.UserHomeDir

// Dir returns ~/.voiceml for the current user. Exported for tests.
func Dir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: %w", err)
	}
	return filepath.Join(home, ".voiceml"), nil
}

// Path returns the absolute path to config.toml.
func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.toml"), nil
}

// HistoryPath returns the absolute path to the readline history file.
func HistoryPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "history"), nil
}

// LoadFrom reads and parses a config file from an explicit path.
// Returns ErrNotFound if the file is missing.
func LoadFrom(path string) (*Config, error) {
	root, err := privateDir(filepath.Dir(path), false)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer func() { _ = root.Close() }()
	name := filepath.Base(path)
	f, err := privateFile(root, name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	data, err := root.ReadFile(name)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err := toml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	return c, nil
}

// Load reads the default ~/.voiceml/config.toml. If the file is
// missing, a Config with defaults is returned alongside ErrNotFound so
// callers can distinguish "first run" from "real error".
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	c, err := LoadFrom(p)
	if errors.Is(err, ErrNotFound) {
		return &Config{BaseURL: DefaultBaseURL}, ErrNotFound
	}
	return c, err
}

// SaveTo writes the config to an explicit path atomically. Creates the
// parent directory with 0700 if missing.
func SaveTo(path string, c *Config) error {
	dir := filepath.Dir(path)
	root, err := privateDir(dir, true)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	if err := regularFile(root, filepath.Base(path)); err != nil && !os.IsNotExist(err) {
		return err
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".config-*.toml.tmp")
	if err != nil {
		return fmt.Errorf("config: temp file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("config: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("config: sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("config: close: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		cleanup()
		return fmt.Errorf("config: chmod: %w", err)
	}
	if err := root.Rename(filepath.Base(tmpName), filepath.Base(path)); err != nil {
		cleanup()
		return fmt.Errorf("config: rename: %w", err)
	}
	return nil
}

// Change contains only fields explicitly chosen for persistence by a command.
type Change struct {
	AccountSid *string
	APIKey     *string
}

// Update merges a command's explicit changes with the on-disk configuration.
// Session-only flags and environment values never enter this configuration.
func Update(change Change) error {
	c, err := Load()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if change.AccountSid != nil {
		c.AccountSid = *change.AccountSid
	}
	if change.APIKey != nil {
		c.APIKey = *change.APIKey
	}
	return Save(c)
}

func privateDir(dir string, create bool) (*os.Root, error) {
	if create {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, err
		}
	}
	st, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("config: directory must be a real directory: %s", dir)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	if err := root.Chmod(".", 0o700); err != nil {
		_ = root.Close()
		return nil, err
	}
	return root, nil
}

func regularFile(root *os.Root, name string) error {
	st, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() {
		return fmt.Errorf("config: refusing non-regular file %s", name)
	}
	return nil
}

func privateFile(root *os.Root, name string) (*os.File, error) {
	if err := regularFile(root, name); err != nil {
		return nil, err
	}
	f, err := root.OpenFile(name, os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}

// PrepareHistory creates private history storage and removes old sensitive
// entries before readline can load them into memory.
func PrepareHistory(path string, keep func(string) bool) error {
	root, err := privateDir(filepath.Dir(path), true)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	name := filepath.Base(path)
	if err := regularFile(root, name); os.IsNotExist(err) {
		f, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		return f.Close()
	} else if err != nil {
		return err
	}
	f, err := privateFile(root, name)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	var clean bytes.Buffer
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); keep(line) {
			clean.WriteString(line)
			clean.WriteByte('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := f.Write(clean.Bytes()); err != nil {
		return err
	}
	return f.Sync()
}

// Save writes the config to ~/.voiceml/config.toml atomically.
func Save(c *Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	return SaveTo(p, c)
}
