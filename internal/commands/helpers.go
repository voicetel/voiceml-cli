package commands

import (
	"encoding/json"
	"fmt"
	"reflect"
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
		return usageError("%s: missing JSON body argument", label)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &fields); err != nil {
		return usageError("%s: invalid JSON object: %w", label, err)
	}
	if fields == nil {
		return usageError("%s: expected a JSON object", label)
	}
	target := reflect.ValueOf(v).Elem()
	known := make(map[string]reflect.StructField)
	for _, field := range reflect.VisibleFields(target.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}
		known[strings.ToLower(field.Name)] = field
		if name := field.Tag.Get("form"); name != "" {
			known[strings.ToLower(name)] = field
		}
	}
	// These wire keys cannot be Go identifiers. Keep the mapping explicit;
	// all other keys come from the SDK's exported fields and form tags.
	for wire, goName := range map[string]string{
		"StartTime<": "StartTimeLt", "StartTime>": "StartTimeGt",
		"EndTime<": "EndTimeLt", "EndTime>": "EndTimeGt",
		"DateCreated<": "DateCreatedLt", "DateCreated>": "DateCreatedGt",
		"DateUpdated<": "DateUpdatedLt", "DateUpdated>": "DateUpdatedGt",
		"DateSent<": "DateSentLess", "DateSent>": "DateSentMore",
	} {
		if field, ok := known[strings.ToLower(goName)]; ok {
			known[strings.ToLower(wire)] = field
		}
	}
	seen := make(map[string]bool)
	for name, raw := range fields {
		field, ok := known[strings.ToLower(name)]
		if !ok {
			return usageError("%s: unknown JSON key %q", label, name)
		}
		if seen[field.Name] {
			return usageError("%s: duplicate aliases for JSON key %q", label, name)
		}
		seen[field.Name] = true
		if err := json.Unmarshal(raw, target.FieldByIndex(field.Index).Addr().Interface()); err != nil {
			return usageError("%s: invalid JSON key %q: %w", label, name, err)
		}
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
		return usageError("%s: expected %d argument(s) — %s", label, n, hint)
	}
	return nil
}

func requireConfigured(c *Context) error {
	if c.Client.AccountSid() == "" || c.Client.APIKey() == "" {
		return fmt.Errorf("credentials not configured — run `login <account_sid>` or set VOICEML_ACCOUNT_SID + VOICEML_API_KEY")
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
	if len(s) >= 16 {
		return "..." + s[len(s)-4:]
	}
	return "<set>"
}
