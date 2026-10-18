// Package output centralises how the CLI prints JSON results and errors.
//
// We pretty-print successful responses with encoding/json's MarshalIndent
// using two-space indentation. When the destination is a TTY we additionally
// ANSI-colour the JSON (cyan keys, green strings, yellow numbers, magenta
// literals) and wrap errors in red. When the writer is not a TTY (logs,
// pipes, redirected output) we drop all escape codes so output stays clean.
package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	voiceml "github.com/voicetel/voiceml-go-sdk"
	"golang.org/x/term"
)

// ANSI escape codes. Only emitted when the destination is a TTY.
const (
	ansiReset   = "\x1b[0m"
	ansiRed     = "\x1b[31m"
	ansiBold    = "\x1b[1m"
	ansiKey     = "\x1b[36m" // cyan
	ansiString  = "\x1b[32m" // green
	ansiNumber  = "\x1b[33m" // yellow
	ansiLiteral = "\x1b[35m" // magenta (true / false / null)
)

// Printer formats command results to a writer. Construct one per
// process; it is safe for concurrent use as long as the underlying
// writer is.
type Printer struct {
	out      io.Writer
	err      io.Writer
	colorOut bool
	colorErr bool
}

// New returns a Printer that writes to out (success) and err (errors).
// Colour is determined independently for each destination and honours NO_COLOR.
func New(out, err io.Writer) *Printer {
	_, noColor := os.LookupEnv("NO_COLOR")
	return &Printer{out: out, err: err, colorOut: !noColor && isTerminal(out), colorErr: !noColor && isTerminal(err)}
}

// NewWithColor lets tests opt in or out of colour explicitly.
func NewWithColor(out, err io.Writer, color bool) *Printer {
	return &Printer{out: out, err: err, colorOut: color, colorErr: color}
}

// JSON pretty-prints v to the success writer. When colour is enabled,
// the output is ANSI-syntax-highlighted (keys / strings / numbers /
// literals get distinct foreground colours).
func (p *Printer) JSON(v any) error {
	if v == nil {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("output: marshal: %w", err)
	}
	b := bytes.TrimSuffix(buf.Bytes(), []byte{'\n'})
	payload := string(b)
	if p.colorOut {
		payload = colorizeJSON(b)
	}
	if _, err := fmt.Fprintln(p.out, payload); err != nil {
		return fmt.Errorf("output: write: %w", err)
	}
	return nil
}

// Println writes a plain message to the success writer with a trailing newline.
func (p *Printer) Println(msg string) {
	_, _ = fmt.Fprintln(p.out, msg)
}

// Printf writes a formatted message to the success writer.
func (p *Printer) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.out, format, args...)
}

// Error writes err to the error writer, colourising in red when the
// writer is a TTY.
func (p *Printer) Error(err error) {
	if err == nil {
		return
	}
	message := err.Error()
	var apiErr *voiceml.APIError
	if errors.As(err, &apiErr) {
		message = fmt.Sprintf("%s (code: %d, more_info: %s)", message, apiErr.Code, apiErr.MoreInfo)
	}
	if p.colorErr {
		_, _ = fmt.Fprintf(p.err, "%s%sError:%s %s\n", ansiRed, ansiBold, ansiReset, message)
		return
	}
	_, _ = fmt.Fprintf(p.err, "Error: %s\n", message)
}

// Errorf writes a formatted error message.
func (p *Printer) Errorf(format string, args ...any) {
	p.Error(fmt.Errorf(format, args...))
}

// Color reports whether colour codes are emitted. Useful for tests.
func (p *Printer) Color() bool { return p.colorOut }

// Stdout returns the underlying success writer. Callers that need to write
// raw bytes (e.g. an EOF newline from the REPL loop) use this rather than
// duplicating an `out` field on every caller's state.
func (p *Printer) Stdout() io.Writer { return p.out }

// Stderr returns the underlying error writer. Same rationale as Stdout.
func (p *Printer) Stderr() io.Writer { return p.err }

// isTerminal returns true when w is *os.File pointing at a terminal.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// colorizeJSON walks well-formed indented JSON bytes (as produced by
// json.MarshalIndent) and wraps tokens in ANSI escape codes:
//
//   - object key strings  → cyan
//   - non-key strings     → green
//   - numbers             → yellow
//   - true / false / null → magenta
//
// Everything else (braces, brackets, commas, colons, whitespace) is
// emitted verbatim. We use a hand-rolled scanner — small, zero allocations
// per character, robust against backslash-escaped quotes inside strings.
func colorizeJSON(b []byte) string {
	var out bytes.Buffer
	out.Grow(len(b) + 16*8) // rough headroom for ~16 ANSI escape pairs
	for i := 0; i < len(b); {
		ch := b[i]
		switch {
		case ch == '"':
			// String token — scan to closing quote, respecting escapes.
			end := i + 1
			for end < len(b) {
				if b[end] == '\\' && end+1 < len(b) {
					end += 2
					continue
				}
				if b[end] == '"' {
					end++
					break
				}
				end++
			}
			str := b[i:end]
			// Lookahead: skip whitespace; if next non-ws byte is ':' it's an
			// object key.
			j := end
			for j < len(b) && (b[j] == ' ' || b[j] == '\t') {
				j++
			}
			if j < len(b) && b[j] == ':' {
				out.WriteString(ansiKey)
			} else {
				out.WriteString(ansiString)
			}
			out.Write(str)
			out.WriteString(ansiReset)
			i = end
		case ch == 't' && hasPrefix(b, i, "true"):
			out.WriteString(ansiLiteral)
			out.WriteString("true")
			out.WriteString(ansiReset)
			i += 4
		case ch == 'f' && hasPrefix(b, i, "false"):
			out.WriteString(ansiLiteral)
			out.WriteString("false")
			out.WriteString(ansiReset)
			i += 5
		case ch == 'n' && hasPrefix(b, i, "null"):
			out.WriteString(ansiLiteral)
			out.WriteString("null")
			out.WriteString(ansiReset)
			i += 4
		case ch == '-' || (ch >= '0' && ch <= '9'):
			end := i
			if ch == '-' {
				end++
			}
			for end < len(b) {
				c := b[end]
				if (c >= '0' && c <= '9') || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-' {
					end++
					continue
				}
				break
			}
			out.WriteString(ansiNumber)
			out.Write(b[i:end])
			out.WriteString(ansiReset)
			i = end
		default:
			out.WriteByte(ch)
			i++
		}
	}
	return out.String()
}

func hasPrefix(b []byte, i int, s string) bool {
	if i+len(s) > len(b) {
		return false
	}
	for k := 0; k < len(s); k++ {
		if b[i+k] != s[k] {
			return false
		}
	}
	return true
}
