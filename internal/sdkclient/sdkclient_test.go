package sdkclient

import (
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

const (
	testUA         = "voiceml-cli-test/0.0.0"
	testAccountSid = "ACffffffffffffffffffffffffffffffff"
	testAPIKey     = "test-api-key-not-real" // #nosec G101 -- test fixture only
	testBaseURL    = "https://example.invalid"
)

func TestNewConstructsClient(t *testing.T) {
	c := New(testBaseURL, testAccountSid, testAPIKey, testUA)
	if got := c.BaseURL(); got != testBaseURL {
		t.Errorf("BaseURL: want %q, got %q", testBaseURL, got)
	}
	if got := c.AccountSid(); got != testAccountSid {
		t.Errorf("AccountSid: want %q, got %q", testAccountSid, got)
	}
	if got := c.APIKey(); got != testAPIKey {
		t.Errorf("APIKey: want %q, got %q", testAPIKey, got)
	}
}

func TestNewWithEmptyCredentials(t *testing.T) {
	c := New(testBaseURL, "", "", testUA)
	if c.Calls() != nil {
		t.Error("Calls() should be nil without credentials")
	}
	if c.BaseURL() != testBaseURL {
		t.Errorf("BaseURL: got %q", c.BaseURL())
	}
}

func TestSetCredentialsRebuilds(t *testing.T) {
	c := New(testBaseURL, "", "", testUA)
	c.SetCredentials(testAccountSid, testAPIKey)
	if c.AccountSid() != testAccountSid || c.APIKey() != testAPIKey {
		t.Error("credentials not updated")
	}
	if c.Calls() == nil {
		t.Error("Calls() nil after SetCredentials")
	}
}

func TestServiceAccessorsAllReturnNonNil(t *testing.T) {
	c := New(testBaseURL, testAccountSid, testAPIKey, testUA)
	cases := map[string]any{
		"Calls":                c.Calls(),
		"Conferences":          c.Conferences(),
		"Queues":               c.Queues(),
		"Applications":         c.Applications(),
		"Recordings":           c.Recordings(),
		"IncomingPhoneNumbers": c.IncomingPhoneNumbers(),
		"Messages":             c.Messages(),
		"Diagnostics":          c.Diagnostics(),
	}
	for name, svc := range cases {
		if svc == nil {
			t.Errorf("%s(): returned nil", name)
		}
	}
}

func TestDefaultBaseURLWhenEmpty(t *testing.T) {
	c := New("", testAccountSid, testAPIKey, testUA)
	if got := c.BaseURL(); got == "" {
		t.Error("expected non-empty default BaseURL")
	}
	if got := c.BaseURL(); got != voiceml.DefaultBaseURL {
		t.Errorf("BaseURL default: got %q want %q", got, voiceml.DefaultBaseURL)
	}
}
