package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voicetel/voiceml-cli/internal/output"
	"github.com/voicetel/voiceml-cli/internal/sdkclient"
)

type testHarness struct {
	srv     *httptest.Server
	client  sdkclient.Client
	out     *bytes.Buffer
	err     *bytes.Buffer
	context *Context
	cfgSet  *bool
	mu      *muxLike
}

type muxLike struct {
	routes map[string]func(http.ResponseWriter, *http.Request)
}

func (m *muxLike) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + " " + r.URL.Path
	if h, ok := m.routes[key]; ok {
		h(w, r)
		return
	}
	http.Error(w, "no stub for "+key, http.StatusNotFound)
}

func newHarness(t *testing.T) *testHarness {
	t.Helper()
	h := &testHarness{
		mu:     &muxLike{routes: map[string]func(http.ResponseWriter, *http.Request){}},
		out:    &bytes.Buffer{},
		err:    &bytes.Buffer{},
		cfgSet: new(bool),
	}
	h.srv = httptest.NewServer(h.mu)
	t.Cleanup(h.srv.Close)
	h.client = sdkclient.New(h.srv.URL, "ACffffffffffffffffffffffffffffffff", "test-api-key", "voiceml-cli-test/0.0.0")
	h.context = &Context{
		Ctx:     context.Background(),
		Client:  h.client,
		Printer: output.NewWithColor(h.out, h.err, false),
		OnConfigChanged: func() {
			*h.cfgSet = true
		},
	}
	return h
}

func (h *testHarness) stub(method, path string, fn func(http.ResponseWriter, *http.Request)) {
	h.mu.routes[method+" "+path] = fn
}

func (h *testHarness) dispatch(line string, registry *Registry) error {
	return runLine(h.context, registry, line)
}

func runLine(c *Context, r *Registry, line string) error {
	tokens, raw, err := simpleTokenize(line)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}
	return r.Dispatch(c, tokens, raw)
}

func simpleTokenize(line string) ([]string, string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, "", nil
	}
	var tokens []string
	var cur strings.Builder
	inQ := byte(0)
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if inQ != 0 {
			if ch == inQ {
				inQ = 0
				continue
			}
			cur.WriteByte(ch)
			continue
		}
		if ch == '"' || ch == '\'' {
			inQ = ch
			continue
		}
		if ch == ' ' || ch == '\t' {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(ch)
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens, line, nil
}

func TestHelpRoot(t *testing.T) {
	h := newHarness(t)
	r := BuildRegistry()
	if err := h.dispatch("help", r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "Resource groups") {
		t.Errorf("expected resource groups in help: %q", h.out.String())
	}
}

func TestWhoamiNoCredentials(t *testing.T) {
	h := newHarness(t)
	h.client.SetCredentials("", "")
	r := BuildRegistry()
	if err := h.dispatch("whoami", r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "accountSid") {
		t.Errorf("whoami output missing accountSid: %s", h.out.String())
	}
}

func TestLoginSavesCredentials(t *testing.T) {
	h := newHarness(t)
	h.client.SetCredentials("", "")
	r := BuildRegistry()
	if err := h.dispatch("login ACffffffffffffffffffffffffffffffff my-api-key", r); err != nil {
		t.Fatal(err)
	}
	if h.client.AccountSid() == "" || h.client.APIKey() == "" {
		t.Error("login did not install credentials")
	}
	if !*h.cfgSet {
		t.Error("OnConfigChanged not invoked")
	}
}

func TestCallsListRequiresCredentials(t *testing.T) {
	h := newHarness(t)
	h.client.SetCredentials("", "")
	r := BuildRegistry()
	if err := h.dispatch("calls list", r); err == nil {
		t.Fatal("expected credentials error")
	}
}

func TestCallsListSuccess(t *testing.T) {
	h := newHarness(t)
	h.stub("GET", "/2010-04-01/Accounts/ACffffffffffffffffffffffffffffffff/Calls.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"calls":     []any{},
			"page":      0,
			"page_size": 50,
		})
	})
	r := BuildRegistry()
	if err := h.dispatch("calls list", r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "calls") {
		t.Errorf("expected calls list output: %q", h.out.String())
	}
}

func TestDiagnosticsHealth(t *testing.T) {
	h := newHarness(t)
	h.stub("GET", "/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	r := BuildRegistry()
	if err := h.dispatch("diagnostics health", r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "ok") {
		t.Errorf("expected health output: %q", h.out.String())
	}
}

func TestCompletionBash(t *testing.T) {
	h := newHarness(t)
	r := BuildRegistry()
	if err := h.dispatch("completion bash", r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.out.String(), "complete -F _voiceml_cli_complete voiceml-cli") {
		t.Error("bash completion script missing complete directive")
	}
}

func TestCompletionDataIncludesCallsList(t *testing.T) {
	r := BuildRegistry()
	_, _, groupSubs := r.CompletionData()
	subs, ok := groupSubs["calls"]
	if !ok {
		t.Fatal("calls group missing from completion data")
	}
	found := false
	for _, s := range subs {
		if s == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("calls list missing from completions")
	}
}

func TestUnknownCommand(t *testing.T) {
	h := newHarness(t)
	r := BuildRegistry()
	if err := h.dispatch("nosuchcommand", r); err == nil {
		t.Fatal("expected unknown command error")
	}
}

func TestExitReturnsErrExit(t *testing.T) {
	h := newHarness(t)
	r := BuildRegistry()
	err := h.dispatch("exit", r)
	if err == nil || !strings.Contains(err.Error(), "exit") {
		t.Fatalf("expected ErrExit, got %v", err)
	}
}
