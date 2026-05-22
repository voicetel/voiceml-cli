package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/voicetel/voiceml-cli/internal/commands"
	"github.com/voicetel/voiceml-cli/internal/config"
	"github.com/voicetel/voiceml-cli/internal/output"
	"github.com/voicetel/voiceml-cli/internal/repl"
	"github.com/voicetel/voiceml-cli/internal/sdkclient"
)

// Environment variable names recognized by the CLI. Flag values take
// precedence over env vars; env vars take precedence over ~/.voiceml/config.toml.
//
// Both VOICEMEL_* and VOICEML_* spellings are accepted for each variable.
//
//nolint:gosec // env-var NAMES (read via os.Getenv), not credential values; G101 false-positives on API_KEY substrings.
const (
	envAccountSid = "VOICEML_ACCOUNT_SID"
	envAPIKey     = "VOICEML_API_KEY"
	envBaseURL    = "VOICEML_BASE_URL"
)

// exitOnErr is the indirection layer between main() and os.Exit. Tests
// override it to assert on the exit code without actually terminating the
// test binary.
var exitOnErr = func(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "voiceml-cli:", err)
		os.Exit(1)
	}
}

func main() {
	exitOnErr(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("voiceml-cli", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		baseURL    = fs.String("base-url", "", "Override the API endpoint for this session.")
		accountSid = fs.String("account-sid", "", "Account SID for this session (not persisted).")
		apiKey     = fs.String("api-key", "", "API key for this session (not persisted).")
		oneShot    = fs.String("x", "", "Run a single command non-interactively and exit. Example: -x 'calls list'")
		cpuProfile = fs.String("cpu-profile", "", "Write a CPU profile to the given file (e.g. cpu.pprof). Hidden debug flag.")
		memProfile = fs.String("mem-profile", "", "Write a heap profile to the given file (e.g. mem.pprof). Hidden debug flag.")
		showVer    = fs.Bool("version", false, "Print version and exit.")
	)
	fs.Usage = func() { usage(stderr, fs) }
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVer {
		_, _ = fmt.Fprintf(stdout, "voiceml-cli %s\n", Version)
		_, _ = fmt.Fprintf(stdout, "  build time: %s\n", BuildTime)
		_, _ = fmt.Fprintf(stdout, "  git commit: %s\n", GitCommit)
		return nil
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			return fmt.Errorf("cpu-profile: %w", err)
		}
		defer func() { _ = f.Close() }()
		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("cpu-profile: %w", err)
		}
		defer pprof.StopCPUProfile()
	}

	printer := output.New(stdout, stderr)

	cfg, err := config.Load()
	if err != nil && !errors.Is(err, config.ErrNotFound) {
		return fmt.Errorf("config: %w", err)
	}

	applyEnv(cfg)
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	}
	if *accountSid != "" {
		cfg.AccountSid = *accountSid
	}
	if *apiKey != "" {
		cfg.APIKey = *apiKey
	}

	client := sdkclient.New(cfg.BaseURL, cfg.AccountSid, cfg.APIKey, "voiceml-cli/"+Version)

	ctx, cancel := signalContext(*oneShot != "")
	defer cancel()

	defer writeMemProfile(*memProfile, stderr)

	if *oneShot != "" {
		return runOneShot(ctx, client, printer, cfg, *oneShot)
	}

	histPath, err := config.HistoryPath()
	if err != nil {
		return err
	}

	printBanner(printer, client)

	return runLoop(ctx, loopOptions{
		Client:   client,
		Printer:  printer,
		HistFile: histPath,
		Prompt:   "voiceml> ",
		Cfg:      cfg,
	})
}

func applyEnv(cfg *config.Config) {
	if env := firstEnv("VOICEMEL_BASE_URL", envBaseURL); env != "" {
		cfg.BaseURL = env
	}
	if env := firstEnv("VOICEMEL_ACCOUNT_SID", envAccountSid); env != "" {
		cfg.AccountSid = env
	}
	if env := firstEnv("VOICEMEL_API_KEY", envAPIKey); env != "" {
		cfg.APIKey = env
	}
}

func firstEnv(names ...string) string {
	for _, n := range names {
		if v := os.Getenv(n); v != "" {
			return v
		}
	}
	return ""
}

func usage(w io.Writer, fs *flag.FlagSet) {
	pf := func(format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }
	pl := func(s string) { _, _ = fmt.Fprintln(w, s) }
	pf("voiceml-cli %s — interactive REPL for the VoiceML REST API.\n\n", Version)
	pl("Usage:")
	pl("  voiceml-cli [--base-url=URL] [--account-sid=SID] [--api-key=KEY]")
	pl("  voiceml-cli -x '<command>'      # one-shot, no REPL")
	pl("")
	pl("Environment variables (override config, are overridden by flags):")
	pf("  VOICEML_ACCOUNT_SID / VOICEMEL_ACCOUNT_SID   Twilio-format account SID (HTTP Basic username)\n")
	pf("  VOICEML_API_KEY / VOICEMEL_API_KEY           Per-tenant API key (HTTP Basic password)\n")
	pf("  VOICEML_BASE_URL / VOICEMEL_BASE_URL         Override the API endpoint\n")
	pl("")
	pl("Inside the REPL, type `help` for every command. Exit with `exit`, `quit`, or Ctrl-D.")
	pl("")
	pl("Flags:")
	fs.PrintDefaults()
}

func writeMemProfile(path string, stderr io.Writer) {
	if path == "" {
		return
	}
	f, err := os.Create(path) //nolint:gosec // user-controlled profile output path
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "mem-profile: %v\n", err)
		return
	}
	defer func() { _ = f.Close() }()
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		_, _ = fmt.Fprintf(stderr, "mem-profile: %v\n", err)
	}
}

func runOneShot(ctx context.Context, client sdkclient.Client, printer *output.Printer, cfg *config.Config, line string) error {
	registry := commands.BuildRegistry()
	cctx := &commands.Context{
		Ctx:     ctx,
		Client:  client,
		Printer: printer,
		OnConfigChanged: func() {
			cfg.AccountSid = client.AccountSid()
			cfg.APIKey = client.APIKey()
			cfg.BaseURL = client.BaseURL()
			if err := config.Save(cfg); err != nil {
				printer.Errorf("config: save: %v", err)
			}
		},
	}
	p, err := repl.Parse(line)
	if err != nil {
		if errors.Is(err, repl.ErrEmpty) {
			return fmt.Errorf("-x: empty command")
		}
		return fmt.Errorf("-x: %w", err)
	}
	if err := registry.Dispatch(cctx, p.Tokens, p.Raw); err != nil {
		if errors.Is(err, commands.ErrExit) {
			return nil
		}
		return err
	}
	return nil
}

func printBanner(p *output.Printer, c sdkclient.Client) {
	p.Printf("VoiceML CLI %s  —  type `help` for commands, `exit` to quit.\n", Version)
	p.Printf("Endpoint: %s\n", c.BaseURL())
	if c.AccountSid() == "" || c.APIKey() == "" {
		p.Printf("No credentials configured. Run `login <account_sid> <api_key>` or set env vars.\n")
	}
}

func signalContext(oneShotMode bool) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := []os.Signal{syscall.SIGTERM}
	if oneShotMode {
		sigs = append(sigs, syscall.SIGINT)
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	go func() {
		<-ch
		cancel()
	}()
	return ctx, cancel
}
