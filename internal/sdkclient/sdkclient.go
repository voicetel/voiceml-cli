// Package sdkclient is a thin abstraction over *voiceml.Client so that
// command dispatch can be tested without making real HTTP calls.
//
// The interface intentionally mirrors only the surface area the CLI uses;
// new commands extend the interface (and the test fake) rather than
// reaching for the concrete *voiceml.Client.
package sdkclient

import (
	voiceml "github.com/voicetel/voiceml-go-sdk"
)

// Client is the minimal surface the CLI relies on. A real *voiceml.Client
// satisfies it via realClient; tests substitute a fake.
type Client interface {
	BaseURL() string
	AccountSid() string
	APIKey() string

	SetCredentials(accountSid, apiKey string)
	SetAccountSid(accountSid string)
	SetAPIKey(apiKey string)

	Calls() *voiceml.CallsService
	Conferences() *voiceml.ConferencesService
	Queues() *voiceml.QueuesService
	Applications() *voiceml.ApplicationsService
	Recordings() *voiceml.RecordingsService
	IncomingPhoneNumbers() *voiceml.IncomingPhoneNumbersService
	Messages() *voiceml.MessagesService
	Diagnostics() *voiceml.DiagnosticsService
}

// realClient adapts *voiceml.Client to the Client interface. Because the
// SDK client is constructed from options, we hold the construction state
// and rebuild when credentials change.
type realClient struct {
	inner      *voiceml.Client
	baseURL    string
	accountSid string
	apiKey     string
	ua         string
}

// New builds a real SDK-backed Client.
func New(baseURL, accountSid, apiKey, userAgent string) Client {
	r := &realClient{baseURL: baseURL, accountSid: accountSid, apiKey: apiKey, ua: userAgent}
	r.rebuild()
	return r
}

func (r *realClient) rebuild() {
	opts := voiceml.ClientOptions{
		AccountSid: r.accountSid,
		APIKey:     r.apiKey,
		BaseURL:    r.baseURL,
		UserAgent:  r.ua,
	}
	if r.accountSid == "" || r.apiKey == "" {
		r.inner = nil
		return
	}
	c, err := voiceml.NewClient(opts)
	if err != nil {
		r.inner = nil
		return
	}
	r.inner = c
	if r.baseURL == "" {
		r.baseURL = c.BaseURL
	}
}

func (r *realClient) innerOrNil() *voiceml.Client { return r.inner }

func (r *realClient) BaseURL() string {
	if r.inner != nil {
		return r.inner.BaseURL
	}
	if r.baseURL != "" {
		return r.baseURL
	}
	return voiceml.DefaultBaseURL
}

func (r *realClient) AccountSid() string { return r.accountSid }
func (r *realClient) APIKey() string     { return r.apiKey }

func (r *realClient) SetCredentials(accountSid, apiKey string) {
	r.accountSid = accountSid
	r.apiKey = apiKey
	r.rebuild()
}

func (r *realClient) SetAccountSid(accountSid string) {
	r.accountSid = accountSid
	r.rebuild()
}

func (r *realClient) SetAPIKey(apiKey string) {
	r.apiKey = apiKey
	r.rebuild()
}

func (r *realClient) Calls() *voiceml.CallsService {
	if c := r.innerOrNil(); c != nil {
		return c.Calls
	}
	return nil
}
func (r *realClient) Conferences() *voiceml.ConferencesService {
	if c := r.innerOrNil(); c != nil {
		return c.Conferences
	}
	return nil
}
func (r *realClient) Queues() *voiceml.QueuesService {
	if c := r.innerOrNil(); c != nil {
		return c.Queues
	}
	return nil
}
func (r *realClient) Applications() *voiceml.ApplicationsService {
	if c := r.innerOrNil(); c != nil {
		return c.Applications
	}
	return nil
}
func (r *realClient) Recordings() *voiceml.RecordingsService {
	if c := r.innerOrNil(); c != nil {
		return c.Recordings
	}
	return nil
}
func (r *realClient) IncomingPhoneNumbers() *voiceml.IncomingPhoneNumbersService {
	if c := r.innerOrNil(); c != nil {
		return c.IncomingPhoneNumbers
	}
	return nil
}
func (r *realClient) Messages() *voiceml.MessagesService {
	if c := r.innerOrNil(); c != nil {
		return c.Messages
	}
	return nil
}
func (r *realClient) Diagnostics() *voiceml.DiagnosticsService {
	if c := r.innerOrNil(); c != nil {
		return c.Diagnostics
	}
	return nil
}
