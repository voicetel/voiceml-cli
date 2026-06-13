# VoiceML CLI

Interactive REPL for the [VoiceML REST API](https://voicetel.com/docs/api/v0.7/voiceml/) — outbound voice, AMD, conferences, queues, and Twilio-compatible resources on top of the [VoiceML Go SDK](../go).

![Version](https://img.shields.io/badge/version-0.2.0-blue)
![Go](https://img.shields.io/badge/go-1.22%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platforms](https://img.shields.io/badge/platforms-linux%20%7C%20macos%20%7C%20windows-lightgrey)

## Features

- **Interactive REPL** with readline editing, persistent history (`~/.voiceml/history`), and tab completion
- **One-shot mode** via `-x 'command'` for scripts and CI
- **HTTP Basic auth** — Account SID (username) + API key (password)
- **Colorized JSON** on TTYs; plain JSON when piped
- **Resource groups**: calls, conferences, queues, applications, recordings, incoming-phone-numbers, messages, diagnostics
- **Cross-compile** to 9 platforms via `make build-all`

## Installation

```bash
cd cli
make build    # → bin/voiceml-cli
make install  # → $GOPATH/bin/voiceml-cli
```

## Quickstart

```text
$ voiceml-cli
VoiceML CLI 0.2.0  —  type `help` for commands, `exit` to quit.
Endpoint: https://voiceml.voicetel.com
No credentials configured. Run `login <account_sid> <api_key>` or set env vars.

voiceml> login ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx your-api-key
Credentials installed and saved.

voiceml> calls list
{
  "calls": [],
  "page": 0,
  ...
}

voiceml> whoami
{
  "accountSid": "ACxx...xxxx",
  "apiKey": "your...-key",
  "baseURL": "https://voiceml.voicetel.com",
  "auth": "HTTP Basic (AccountSid username + API key password)"
}
```

## One-shot mode

```bash
export VOICEML_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
export VOICEML_API_KEY=your-api-key
voiceml-cli -x 'calls list'
voiceml-cli -x 'diagnostics health'
```

## Environment variables

Precedence: **flag > env > config file**.

| Variable | Alternate | Purpose |
|---|---|---|
| `VOICEML_ACCOUNT_SID` | `VOICEMEL_ACCOUNT_SID` | Account SID (HTTP Basic username) |
| `VOICEML_API_KEY` | `VOICEMEL_API_KEY` | API key (HTTP Basic password) |
| `VOICEML_BASE_URL` | `VOICEMEL_BASE_URL` | API endpoint override |

## Configuration

`~/.voiceml/config.toml`:

```toml
account_sid = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
api_key     = "your-api-key"
base_url    = "https://voiceml.voicetel.com"
```

## Building

```bash
make build       # local platform
make build-all   # 9 platforms → ./dist/
make release     # build-all + .tar.gz / .zip archives
make test
```

## Command reference

```text
help [topic]                         Show help
exit | quit                          Leave the REPL
login <account_sid> <api_key>        Save HTTP Basic credentials
set account-sid|api-key <value>      Update credentials
whoami                               Show redacted credentials
completion bash|zsh                  Shell completion script
clear                                Clear the screen

calls list [json]                    List calls (filters: To, From, Status, Page, PageSize, …)
calls get <sid>
calls create <json>
calls update <sid> <json>
calls delete <sid>
calls start-payment <call_sid> <json>
calls update-payment <call_sid> <payment_sid> [json]

conferences list|get|end|list-participants|get-participant|update-participant|kick-participant|list-recordings

queues list|get|create|update|delete|list-members|peek-front|dequeue-front|get-member|dequeue-member

applications list|get|create|update|delete

recordings list [json]|get|get-audio|delete

incoming-phone-numbers list [json]|get|create|update|delete

messages list [json]|get|create|update|delete

diagnostics health|openapi
```

List commands accept optional JSON filter objects matching the Go SDK list params (v0.7.x), e.g.:

```text
calls list {"Status":"completed","PageSize":50}
incoming-phone-numbers list {"PhoneNumber":"+18005551234"}
recordings list {"PageSize":20}
```

## License

MIT — see [LICENSE](LICENSE).
