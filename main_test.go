package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chzyer/readline"

	"github.com/voicetel/voiceml-cli/internal/commands"
	"github.com/voicetel/voiceml-cli/internal/config"
	"github.com/voicetel/voiceml-cli/internal/output"
	"github.com/voicetel/voiceml-cli/internal/repl"
	"github.com/voicetel/voiceml-cli/internal/sdkclient"
)

type scriptedSource struct {
	lines []string
	i     int
	err   error
}

func (s *scriptedSource) Readline() (string, error) {
	if s.i >= len(s.lines) {
		if s.err != nil {
			return "", s.err
		}
		return "", io.EOF
	}
	line := s.lines[s.i]
	s.i++
	return line, nil
}

func (s *scriptedSource) Close() error { return nil }

func isolatedHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	return tmp
}

func TestVersionVarsHaveValues(t *testing.T) {
	if Version == "" || BuildTime == "" || GitCommit == "" {
		t.Error("version vars should have defaults")
	}
}

func TestSignalContextLifecycle(t *testing.T) {
	ctx, cancel := signalContext(false)
	defer cancel()
	if ctx.Err() != nil {
		t.Error("ctx cancelled at construction")
	}
	cancel()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Error("ctx not cancelled")
	}
}

func TestPrintBannerVariants(t *testing.T) {
	var stdout, stderr bytes.Buffer
	p := output.New(&stdout, &stderr)
	c := sdkclient.New("https://example.invalid", "ACffffffffffffffffffffffffffffffff", "key", "test-ua/0.0")
	printBanner(p, c)
	if !strings.Contains(stdout.String(), "VoiceML CLI") {
		t.Errorf("banner missing product name: %q", stdout.String())
	}

	stdout.Reset()
	c = sdkclient.New("https://example.invalid", "", "", "test-ua/0.0")
	printBanner(p, c)
	if !strings.Contains(stdout.String(), "No credentials configured") {
		t.Errorf("expected credentials hint: %q", stdout.String())
	}
}

func TestRunOneShotHelp(t *testing.T) {
	isolatedHome(t)
	var stdout, stderr bytes.Buffer
	cfg := &config.Config{BaseURL: "https://example.invalid", AccountSid: "ACx", APIKey: "k"}
	client := sdkclient.New(cfg.BaseURL, cfg.AccountSid, cfg.APIKey, "test-ua/0.0")
	if err := runOneShot(context.Background(), client, output.New(&stdout, &stderr), cfg, "help"); err != nil {
		t.Fatalf("runOneShot(help): %v", err)
	}
	if !strings.Contains(stdout.String(), "Resource groups") {
		t.Errorf("expected help output, got %q", stdout.String())
	}
}

func TestRunOneShotEmpty(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cfg := &config.Config{BaseURL: "https://example.invalid"}
	client := sdkclient.New(cfg.BaseURL, "ACx", "k", "test-ua/0.0")
	err := runOneShot(context.Background(), client, output.New(&stdout, &stderr), cfg, "   ")
	if err == nil || !strings.Contains(err.Error(), "empty command") {
		t.Fatalf("expected empty command error, got %v", err)
	}
}

func TestRunVersionFlag(t *testing.T) {
	isolatedHome(t)
	var stdout, stderr bytes.Buffer
	if err := run([]string{"--version"}, &stdout, &stderr); err != nil {
		t.Fatalf("run --version: %v", err)
	}
	if !strings.Contains(stdout.String(), "voiceml-cli") {
		t.Errorf("--version output: %q", stdout.String())
	}
}

func TestRunBadFlagReturnsError(t *testing.T) {
	isolatedHome(t)
	var stdout, stderr bytes.Buffer
	if err := run([]string{"--nosuchflag"}, &stdout, &stderr); err == nil {
		t.Fatal("expected error for bad flag")
	}
}

func TestRunOneShotViaFlagX(t *testing.T) {
	isolatedHome(t)
	var stdout, stderr bytes.Buffer
	if err := run([]string{
		"--account-sid=ACffffffffffffffffffffffffffffffff",
		"--api-key=test-key",
		"-x", "help",
	}, &stdout, &stderr); err != nil {
		t.Fatalf("run -x help: %v", err)
	}
	if !strings.Contains(stdout.String(), "Resource groups") {
		t.Errorf("expected help output, got %q", stdout.String())
	}
}

func TestRunEnvVars(t *testing.T) {
	isolatedHome(t)
	t.Setenv("VOICEMEL_ACCOUNT_SID", "ACffffffffffffffffffffffffffffffff")
	t.Setenv("VOICEMEL_API_KEY", "env-key")
	t.Setenv("VOICEMEL_BASE_URL", "https://staging.example.invalid")
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-x", "help"}, &stdout, &stderr); err != nil {
		t.Fatalf("run with env: %v", err)
	}
}

func TestRunLoopWithEOF(t *testing.T) {
	isolatedHome(t)
	var stdout, stderr, eofBuf bytes.Buffer
	cfg := &config.Config{BaseURL: "https://example.invalid", AccountSid: "ACx", APIKey: "k"}
	client := sdkclient.New(cfg.BaseURL, cfg.AccountSid, cfg.APIKey, "test-ua/0.0")
	opts := loopOptions{Client: client, Printer: output.New(&stdout, &stderr), Cfg: cfg}
	registry := commands.BuildRegistry()
	src := &scriptedSource{lines: []string{"help"}}
	if err := runLoopWith(context.Background(), opts, registry, src, &eofBuf); err != nil {
		t.Fatalf("runLoopWith: %v", err)
	}
	if !strings.Contains(stdout.String(), "Resource groups") {
		t.Error("help didn't dispatch in loop")
	}
}

func TestRunLoopWithExit(t *testing.T) {
	isolatedHome(t)
	var stdout, stderr, eofBuf bytes.Buffer
	cfg := &config.Config{BaseURL: "https://example.invalid", AccountSid: "ACx", APIKey: "k"}
	client := sdkclient.New(cfg.BaseURL, cfg.AccountSid, cfg.APIKey, "test-ua/0.0")
	opts := loopOptions{Client: client, Printer: output.New(&stdout, &stderr), Cfg: cfg}
	registry := commands.BuildRegistry()
	src := &scriptedSource{lines: []string{"exit"}}
	if err := runLoopWith(context.Background(), opts, registry, src, &eofBuf); err != nil {
		t.Fatalf("runLoopWith(exit): %v", err)
	}
}

func TestDispatchLineEmpty(t *testing.T) {
	registry := commands.BuildRegistry()
	cctx := &commands.Context{Ctx: context.Background(), Printer: output.New(io.Discard, io.Discard)}
	if err := dispatchLine(cctx, registry, ""); err != nil {
		t.Errorf("expected nil for empty line, got %v", err)
	}
}

func TestUsagePrints(t *testing.T) {
	var buf bytes.Buffer
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("x", "", "")
	usage(&buf, fs)
	got := buf.String()
	for _, want := range []string{"voiceml-cli", envAccountSid, envAPIKey, envBaseURL, "VOICEMEL"} {
		if !strings.Contains(got, want) {
			t.Errorf("usage missing %q", want)
		}
	}
}

func TestRunOneShotConfigCallbackSaves(t *testing.T) {
	tmp := isolatedHome(t)
	var stdout, stderr bytes.Buffer
	cfg := &config.Config{BaseURL: "https://example.invalid"}
	client := sdkclient.New(cfg.BaseURL, "", "", "test-ua/0.0")
	if err := runOneShot(context.Background(), client, output.New(&stdout, &stderr), cfg, "login ACffffffffffffffffffffffffffffffff secret"); err != nil {
		t.Fatalf("login: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".voiceml", "config.toml")); err != nil {
		t.Errorf("config.toml not written: %v", err)
	}
}

type interruptOnceSource struct {
	lines   []string
	i       int
	tripped bool
}

func (s *interruptOnceSource) Readline() (string, error) {
	if !s.tripped {
		s.tripped = true
		return "", readline.ErrInterrupt
	}
	if s.i >= len(s.lines) {
		return "", io.EOF
	}
	line := s.lines[s.i]
	s.i++
	return line, nil
}

func (s *interruptOnceSource) Close() error { return nil }

func TestRunLoopWithInterruptContinues(t *testing.T) {
	isolatedHome(t)
	var stdout, stderr, eofBuf bytes.Buffer
	cfg := &config.Config{BaseURL: "https://example.invalid", AccountSid: "ACx", APIKey: "k"}
	client := sdkclient.New(cfg.BaseURL, cfg.AccountSid, cfg.APIKey, "test-ua/0.0")
	opts := loopOptions{Client: client, Printer: output.New(&stdout, &stderr), Cfg: cfg}
	registry := commands.BuildRegistry()
	src := &interruptOnceSource{lines: []string{"help"}}
	if err := runLoopWith(context.Background(), opts, registry, src, &eofBuf); err != nil {
		t.Fatalf("runLoopWith(interrupt): %v", err)
	}
}

func TestMainBinaryPrintsVersion(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "voiceml-cli-test")
	cmd := exec.Command("go", "build", "-o", bin, ".") //nolint:gosec
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	cmd = exec.Command(bin, "--version") //nolint:gosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "voiceml-cli") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestPackageImports(t *testing.T) {
	if _, err := repl.Parse("noop"); err != nil {
		t.Errorf("repl.Parse roundtrip: %v", err)
	}
}
