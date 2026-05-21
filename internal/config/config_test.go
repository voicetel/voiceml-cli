package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromMissing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nope.toml")
	_, err := LoadFrom(p)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	in := &Config{
		AccountSid: "ACffffffffffffffffffffffffffffffff",
		APIKey:     "secret-api-key",
		BaseURL:    "https://staging.voiceml.example",
	}
	if err := SaveTo(p, in); err != nil {
		t.Fatalf("save: %v", err)
	}
	out, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out.AccountSid != in.AccountSid {
		t.Errorf("account sid: got %q want %q", out.AccountSid, in.AccountSid)
	}
	if out.APIKey != in.APIKey {
		t.Errorf("api key: got %q want %q", out.APIKey, in.APIKey)
	}
	if out.BaseURL != in.BaseURL {
		t.Errorf("base url: got %q want %q", out.BaseURL, in.BaseURL)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := st.Mode().Perm(); mode != 0o600 {
		t.Errorf("mode = %o, want 0600", mode)
	}
}

func TestLoadFromDefaultsBaseURL(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(p, []byte("api_key = \"abc\"\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	out, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out.BaseURL != DefaultBaseURL {
		t.Errorf("base url default: got %q want %q", out.BaseURL, DefaultBaseURL)
	}
}

func TestDirUsesHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	d, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(d) != ".voiceml" {
		t.Errorf("dir = %q, want suffix .voiceml", d)
	}
}
