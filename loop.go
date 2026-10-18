package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/chzyer/readline"

	"github.com/voicetel/voiceml-cli/internal/commands"
	"github.com/voicetel/voiceml-cli/internal/config"
	"github.com/voicetel/voiceml-cli/internal/output"
	"github.com/voicetel/voiceml-cli/internal/repl"
	"github.com/voicetel/voiceml-cli/internal/sdkclient"
)

type loopOptions struct {
	Client   sdkclient.Client
	Printer  *output.Printer
	HistFile string
	Prompt   string
	Cfg      *config.Config
}

type lineSource interface {
	Readline() (string, error)
	Close() error
}

func runLoop(ctx context.Context, opts loopOptions) error {
	registry := commands.BuildRegistry()
	builtins, helpTopics, groupSubs := registry.CompletionData()
	if err := config.PrepareHistory(opts.HistFile, repl.RecordHistory); err != nil {
		return fmt.Errorf("repl: history: %w", err)
	}
	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 opts.Prompt,
		HistoryFile:            opts.HistFile,
		AutoComplete:           repl.BuildCompleter(builtins, helpTopics, groupSubs),
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		HistorySearchFold:      true,
		DisableAutoSaveHistory: true,
	})
	if err != nil {
		return fmt.Errorf("repl: init readline: %w", err)
	}
	return runLoopWith(ctx, opts, registry, rl, opts.Printer.Stdout())
}

func runLoopWith(ctx context.Context, opts loopOptions, registry *commands.Registry, src lineSource, eofOut io.Writer) error {
	defer func() { _ = src.Close() }()

	cmdCtx := &commands.Context{
		Ctx:             ctx,
		Client:          opts.Client,
		Printer:         opts.Printer,
		OnConfigChanged: config.Update,
		ReadSecret:      readSecret,
	}

	for {
		line, err := src.Readline()
		if errors.Is(err, readline.ErrInterrupt) {
			continue
		}
		if errors.Is(err, io.EOF) {
			_, err := fmt.Fprintln(eofOut)
			return err
		}
		if err != nil {
			return fmt.Errorf("repl: read line: %w", err)
		}
		if history, ok := src.(interface{ SaveHistory(string) error }); ok && repl.RecordHistory(line) {
			if err := history.SaveHistory(line); err != nil {
				return fmt.Errorf("repl: save history: %w", err)
			}
		}
		if err := dispatchLine(cmdCtx, registry, line); err != nil {
			if errors.Is(err, commands.ErrExit) {
				return nil
			}
			opts.Printer.Error(err)
		}
	}
}

func dispatchLine(cctx *commands.Context, r *commands.Registry, line string) error {
	p, err := repl.Parse(line)
	if err != nil {
		if errors.Is(err, repl.ErrEmpty) {
			return nil
		}
		return err
	}
	return r.Dispatch(cctx, p.Tokens, p.Raw)
}
