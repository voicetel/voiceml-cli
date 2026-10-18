package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/voicetel/voiceml-cli/internal/commands"
	"github.com/voicetel/voiceml-cli/internal/config"
	"github.com/voicetel/voiceml-cli/internal/output"
	"github.com/voicetel/voiceml-cli/internal/repl"
	"github.com/voicetel/voiceml-cli/internal/sdkclient"
)

// Environment variable names recognized by the CLI. Flag values take
// precedence over env vars; env vars take precedence over ~/.voiceml/config.toml.
//
// Canonical VOICEML_* variables win over the legacy VOICEMEL_* aliases.
const (
	envAccountSid = "VOICEML_ACCOUNT_SID"
	envKeyVar     = "VOICEML_API_KEY"
	envBaseURL    = "VOICEML_BASE_URL"
)

// exitOnErr is the indirection layer between main() and os.Exit. Tests
// override it to assert on the exit code without actually terminating the
// test binary.
var exitOnErr = func(err error) {
	if err != nil {
		output.New(os.Stdout, os.Stderr).Error(err)
		os.Exit(exitCode(err))
	}
}

func main() {
	exitOnErr(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) error {
	return runWithInput(args, os.Stdin, stdout, stderr)
}

func exitCode(err error) int {
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return 0
	}
	var usageErr *commands.UsageError
	if errors.As(err, &usageErr) {
		return 2
	}
	return 1
}

func runWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("voiceml-cli", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		baseURL    = fs.String("base-url", "", "Override the API endpoint for this session.")
		accountSid = fs.String("account-sid", "", "Account SID for this session (not persisted).")
		apiKey     = fs.String("api-key", "", "API key for this session (not persisted).")
		oneShot    = fs.String("x", "", "Run a single command non-interactively and exit. Example: -x 'calls list'")
		keyStdin   = fs.Bool("api-key-stdin", false, "Read an ephemeral API key from stdin (one line).")
		showVer    = fs.Bool("version", false, "Print version and exit.")
	)
	fs.Usage = func() { usage(stderr, fs) }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return &commands.UsageError{Err: err}
	}
	if fs.NArg() != 0 {
		return &commands.UsageError{Err: fmt.Errorf("unexpected positional arguments; use -x '<command>'")}
	}
	if *keyStdin && *apiKey != "" {
		return &commands.UsageError{Err: fmt.Errorf("--api-key and --api-key-stdin are mutually exclusive")}
	}

	if *showVer {
		_, _ = fmt.Fprintf(stdout, "voiceml-cli %s\n", Version)
		_, _ = fmt.Fprintf(stdout, "  build time: %s\n", BuildTime)
		_, _ = fmt.Fprintf(stdout, "  git commit: %s\n", GitCommit)
		return nil
	}

	printer := output.New(stdout, stderr)

	cfg, err := config.Load()
	if err != nil && !errors.Is(err, config.ErrNotFound) {
		return fmt.Errorf("config: %w", err)
	}

	applyEnv(cfg, stderr)
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	}
	if *accountSid != "" {
		cfg.AccountSid = *accountSid
	}
	if *apiKey != "" {
		cfg.APIKey = *apiKey
	}
	if *keyStdin {
		// Bound input and require EOF so an accidental multi-line secret is rejected.
		data, err := io.ReadAll(io.LimitReader(stdin, 64*1024+1))
		if err != nil {
			return fmt.Errorf("api-key-stdin: %w", err)
		}
		key := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
		if len(data) > 64*1024 || key == "" || strings.ContainsAny(key, "\r\n") {
			return &commands.UsageError{Err: fmt.Errorf("api-key-stdin: expected one nonempty line, at most 64 KiB")}
		}
		cfg.APIKey = key
	}

	client := sdkclient.New(cfg.BaseURL, cfg.AccountSid, cfg.APIKey, "voiceml-cli/"+Version)

	ctx, cancel := signalContext(*oneShot != "")
	defer cancel()

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

func applyEnv(cfg *config.Config, stderr io.Writer) {
	if env := firstEnv(stderr, envBaseURL, "VOICEMEL_BASE_URL"); env != "" {
		cfg.BaseURL = env
	}
	if env := firstEnv(stderr, envAccountSid, "VOICEMEL_ACCOUNT_SID"); env != "" {
		cfg.AccountSid = env
	}
	if env := firstEnv(stderr, envKeyVar, "VOICEMEL_API_KEY"); env != "" {
		cfg.APIKey = env
	}
}

func firstEnv(stderr io.Writer, canonical, legacy string) string {
	if v := os.Getenv(canonical); v != "" {
		return v
	}
	if v := os.Getenv(legacy); v != "" {
		_, _ = fmt.Fprintf(stderr, "%s is deprecated; use %s instead.\n", legacy, canonical)
		return v
	}
	return ""
}

func usage(w io.Writer, fs *flag.FlagSet) {
	pf := func(format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }
	pl := func(s string) { _, _ = fmt.Fprintln(w, s) }
	pf("voiceml-cli %s — interactive REPL for the VoiceML REST API.\n\n", Version)
	pl("Usage:")
	pl("  voiceml-cli [--base-url=URL] [--account-sid=SID] [--api-key-stdin]")
	pl("  voiceml-cli -x '<command>'      # one-shot, no REPL")
	pl("")
	pl("Environment variables (override config, are overridden by flags):")
	pf("  VOICEML_ACCOUNT_SID / VOICEMEL_ACCOUNT_SID   VoiceML Account SID (HTTP Basic username)\n")
	pf("  VOICEML_API_KEY / VOICEMEL_API_KEY           Per-tenant API key (HTTP Basic password)\n")
	pf("  VOICEML_BASE_URL / VOICEMEL_BASE_URL         Override the API endpoint\n")
	pl("  Canonical VOICEML_* values take precedence over legacy VOICEMEL_* aliases.")
	pl("")
	pl("Inside the REPL, type `help` for every command. Exit with `exit`, `quit`, or Ctrl-D.")
	pl("")
	pl("Flags:")
	fs.PrintDefaults()
}

func readSecret() (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("login: hidden input requires a terminal; use --api-key-stdin for a session key")
	}
	_, _ = fmt.Fprint(os.Stderr, "API key: ")
	key, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("login: read API key: %w", err)
	}
	return string(key), nil
}

func runOneShot(ctx context.Context, client sdkclient.Client, printer *output.Printer, _ *config.Config, line string) error {
	registry := commands.BuildRegistry()
	cctx := &commands.Context{
		Ctx:             ctx,
		Client:          client,
		Printer:         printer,
		OnConfigChanged: config.Update,
		ReadSecret:      readSecret,
	}
	p, err := repl.Parse(line)
	if err != nil {
		if errors.Is(err, repl.ErrEmpty) {
			return &commands.UsageError{Err: fmt.Errorf("-x: empty command")}
		}
		return &commands.UsageError{Err: fmt.Errorf("-x: %w", err)}
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
		p.Printf("No credentials configured. Run `login <account_sid>` for hidden key input or set env vars.\n")
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
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(ch)
		cancel()
	}
}
