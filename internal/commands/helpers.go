package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func skipArgs(tail string, n int) string {
	s := strings.TrimSpace(tail)
	for i := 0; i < n; i++ {
		s = strings.TrimSpace(s)
		end := 0
		inQ := byte(0)
		for end < len(s) {
			ch := s[end]
			if inQ != 0 {
				if ch == inQ {
					inQ = 0
				}
				end++
				continue
			}
			if ch == '"' || ch == '\'' {
				inQ = ch
				end++
				continue
			}
			if unicode.IsSpace(rune(ch)) {
				break
			}
			end++
		}
		s = s[end:]
	}
	return strings.TrimSpace(s)
}

func parseJSON(label, body string, v any) error {
	if body == "" {
		return fmt.Errorf("%s: missing JSON body argument", label)
	}
	if err := json.Unmarshal([]byte(body), v); err != nil {
		return fmt.Errorf("%s: invalid JSON body: %w", label, err)
	}
	return nil
}

func parseOptionalJSON(label, body string, v any) error {
	if strings.TrimSpace(body) == "" {
		return nil
	}
	return parseJSON(label, body, v)
}

func requireArgs(label string, args []string, n int, hint string) error {
	if len(args) < n {
		return fmt.Errorf("%s: expected %d argument(s) — %s", label, n, hint)
	}
	return nil
}

func argInt(label string, raw string) (int, error) {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: expected integer, got %q", label, raw)
	}
	return n, nil
}

func requireConfigured(c *Context) error {
	if c.Client.AccountSid() == "" || c.Client.APIKey() == "" {
		return fmt.Errorf("credentials not configured — run `login <account_sid> <api_key>` or set VOICEML_ACCOUNT_SID + VOICEML_API_KEY")
	}
	return nil
}

func printResultG[T any](c *Context, v *T, err error) error {
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	return c.Printer.JSON(v)
}

func maskSecret(s string) string {
	if s == "" {
		return "<not set>"
	}
	if len(s) >= 8 {
		return s[:4] + "..." + s[len(s)-4:]
	}
	return "<set>"
}
