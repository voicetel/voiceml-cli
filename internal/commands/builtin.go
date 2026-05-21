package commands

import (
	"fmt"
	"strings"
)

func loginCommand() *Command {
	return &Command{
		Names:    []string{"login"},
		Synopsis: "Install Account SID + API key (HTTP Basic auth).",
		Usage: "login <account_sid> <api_key>\n" +
			"  AccountSid is the Twilio-format identifier (AC…). API key is the HTTP Basic password.\n" +
			"  Both values are persisted to ~/.voiceml/config.toml.",
		Run: func(c *Context, args []string, _ string) error {
			if len(args) != 2 {
				return fmt.Errorf("login: expected 2 arguments (account_sid, api_key), got %d", len(args))
			}
			c.Client.SetCredentials(args[0], args[1])
			if c.OnConfigChanged != nil {
				c.OnConfigChanged()
			}
			c.Printer.Println("Credentials installed and saved.")
			return nil
		},
	}
}

func setCommand() *Command {
	return &Command{
		Names:    []string{"set"},
		Synopsis: "Install credentials or override the base URL.",
		Usage: "set account-sid <sid>   Install the Account SID.\n" +
			"set api-key <key>       Install the API key.\n" +
			"set base-url <url>      Override the API endpoint (requires restart).",
		Run: func(c *Context, args []string, _ string) error {
			if len(args) < 2 {
				return fmt.Errorf("set: expected `account-sid`, `api-key`, or `base-url` with a value")
			}
			switch args[0] {
			case "account-sid", "accountsid", "sid":
				c.Client.SetAccountSid(args[1])
				if c.OnConfigChanged != nil {
					c.OnConfigChanged()
				}
				c.Printer.Println("Account SID installed.")
			case "api-key", "apikey", "key":
				c.Client.SetAPIKey(args[1])
				if c.OnConfigChanged != nil {
					c.OnConfigChanged()
				}
				c.Printer.Println("API key installed.")
			case "base-url", "baseurl", "url":
				return fmt.Errorf("set base-url: not mutable after start; restart with --base-url")
			default:
				return fmt.Errorf("set: unknown key %q (use account-sid, api-key, or base-url)", args[0])
			}
			return nil
		},
	}
}

func whoamiCommand() *Command {
	return &Command{
		Names:    []string{"whoami"},
		Synopsis: "Show current Account SID, API key, and base URL.",
		Usage:    "whoami\n  Prints redacted credentials and the configured endpoint.",
		Run: func(c *Context, _ []string, _ string) error {
			return c.Printer.JSON(map[string]any{
				"accountSid": maskSecret(c.Client.AccountSid()),
				"apiKey":     maskSecret(c.Client.APIKey()),
				"baseURL":    c.Client.BaseURL(),
				"auth":       "HTTP Basic (AccountSid username + API key password)",
			})
		},
	}
}

func completionCommand(r *Registry) *Command {
	return &Command{
		Names:    []string{"completion"},
		Synopsis: "Print shell completion script.",
		Usage: "completion bash | zsh\n" +
			"  Eval the output in your shell init file, e.g.:\n" +
			"    eval \"$(voiceml-cli -x 'completion bash')\"",
		Run: func(c *Context, args []string, _ string) error {
			if len(args) != 1 {
				return fmt.Errorf("completion: expected `bash` or `zsh`")
			}
			switch args[0] {
			case "bash":
				c.Printer.Println(renderBashCompletion(r))
			case "zsh":
				c.Printer.Println(renderZshCompletion(r))
			default:
				return fmt.Errorf("completion: unsupported shell %q (use bash or zsh)", args[0])
			}
			return nil
		},
	}
}

func renderBashCompletion(r *Registry) string {
	var b strings.Builder
	b.WriteString(`_voiceml_cli_complete() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"
  local words=(${COMP_WORDS[@]:1:$COMP_CWORD})
  local cmd=""
  local sub=""
  if ((${#words[@]} >= 1)); then cmd="${words[0]}"; fi
  if ((${#words[@]} >= 2)); then sub="${words[1]}"; fi
  local opts=""
`)
	b.WriteString("  case \"${cmd}\" in\n")
	for _, builtin := range r.Builtins {
		name := builtin.Names[0]
		if name == "help" {
			continue
		}
		fmt.Fprintf(&b, "    %s) opts=\"%s\" ;;\n", name, strings.Join(builtin.Names, " "))
	}
	fmt.Fprintf(&b, "    help) opts=\"%s\" ;;\n", strings.Join(completionHelpTopics(r), " "))
	for _, g := range r.Groups {
		subs := make([]string, 0, len(g.Commands))
		for _, cmd := range g.Commands {
			subs = append(subs, cmd.Names[0])
		}
		fmt.Fprintf(&b, "    %s)\n", g.Name)
		fmt.Fprintf(&b, "      if [[ \"${sub}\" == \"\" ]]; then opts=\"%s\"; else opts=\"\"; fi ;;\n", strings.Join(subs, " "))
	}
	b.WriteString(`    *) opts="`)
	top := completionTopLevel(r)
	b.WriteString(strings.Join(top, " "))
	b.WriteString(`" ;;
  esac
  COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
  return 0
}
complete -F _voiceml_cli_complete voiceml-cli
`)
	return b.String()
}

func renderZshCompletion(r *Registry) string {
	var b strings.Builder
	b.WriteString("#compdef voiceml-cli\n\n")
	b.WriteString("_voiceml_cli() {\n  local -a commands groups\n")
	b.WriteString("  commands=(\n")
	for _, name := range completionTopLevel(r) {
		fmt.Fprintf(&b, "    '%s'\n", name)
	}
	b.WriteString("  )\n")
	for _, g := range r.Groups {
		fmt.Fprintf(&b, "  _arguments \"1: :->cmd\" \"2: :->sub\" && case $state in\n")
		fmt.Fprintf(&b, "    (cmd) _describe command commands ;;\n")
		fmt.Fprintf(&b, "    (sub) case $words[1] in\n")
		fmt.Fprintf(&b, "      %s) _values subcommand", g.Name)
		for _, cmd := range g.Commands {
			fmt.Fprintf(&b, " '%s'", cmd.Names[0])
		}
		b.WriteString(" ;;\n")
	}
	b.WriteString("    esac ;;\n  esac\n}\n\n_voiceml_cli\n")
	return b.String()
}

func completionTopLevel(r *Registry) []string {
	out := make([]string, 0, len(r.Builtins)+len(r.Groups))
	for _, b := range r.Builtins {
		out = append(out, b.Names[0])
	}
	for _, g := range r.Groups {
		out = append(out, g.Name)
	}
	return out
}

func completionHelpTopics(r *Registry) []string {
	_, topics, _ := r.CompletionData()
	return topics
}
