package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// UmamiClient sends page-level analytics events to a self-hosted Umami
// instance. Used only in web mode — CLI and TUI telemetry goes through OTel.
type UmamiClient struct {
	endpoint   string // e.g. "http://umami:3000"
	websiteID  string
	hostname   string // e.g. "riddellious.dev" — from UMAMI_HOSTNAME env
	enabled    bool
	httpClient *http.Client
}

// NewUmami creates a client. If endpoint or websiteID is empty, the client
// is disabled (all methods are no-ops).
func NewUmami(endpoint, websiteID, hostname string) *UmamiClient {
	if hostname == "" {
		hostname = "localhost"
	}
	return &UmamiClient{
		endpoint:  endpoint,
		websiteID: websiteID,
		hostname:  hostname,
		enabled:   endpoint != "" && websiteID != "",
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Track sends a named event to Umami. Non-blocking (fires a goroutine).
// Safe to call on a disabled client.
func (u *UmamiClient) Track(event, pageURL string, props map[string]string) {
	if !u.enabled {
		debugf("umami: disabled, skipping event=%s", event)
		return
	}
	debugf("umami: tracking event=%s url=%s", event, pageURL)

	payload := umamiPayload{
		Payload: umamiEvent{
			Hostname: u.hostname,
			Language: "en",
			URL:      pageURL,
			Website:  u.websiteID,
			Name:     event,
			Data:     props,
		},
	}

	go u.send(payload) //nolint:errcheck // fire-and-forget; telemetry errors must not affect tool operation
}

func (u *UmamiClient) send(payload umamiPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, u.endpoint+"/api/send", bytes.NewReader(body)) //nolint:noctx // fire-and-forget telemetry POST; no cancellation needed
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		debugf("umami: send error: %v", err)
		return
	}
	debugf("umami: event=%s status=%d", payload.Payload.Name, resp.StatusCode)
	_ = resp.Body.Close()
}

type umamiPayload struct {
	Payload umamiEvent `json:"payload"`
}

type umamiEvent struct {
	Hostname string            `json:"hostname"`
	Language string            `json:"language"`
	URL      string            `json:"url"`
	Website  string            `json:"website"`
	Name     string            `json:"name"`
	Data     map[string]string `json:"data,omitempty"`
}
