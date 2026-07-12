package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	maxRetries    = 3
	retryBaseWait = 1 * time.Second
	httpTimeout   = 10 * time.Second
)

// Sender dispatches event payloads to an n8n webhook endpoint asynchronously.
type Sender struct {
	url     string
	secret  string
	enabled bool
	client  *http.Client
}

// New creates a new Sender. If enabled is false, Send is a no-op.
func New(url, secret string, enabled bool) *Sender {
	return &Sender{
		url:     url,
		secret:  secret,
		enabled: enabled,
		client: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// Send dispatches the payload asynchronously with retries. Non-blocking.
func (s *Sender) Send(payload any) {
	if !s.enabled || s.url == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		body, err := json.Marshal(payload)
		if err != nil {
			log.Error().Err(err).Msg("webhook: marshal payload failed")
			return
		}

		for attempt := 1; attempt <= maxRetries; attempt++ {
			if err := s.post(ctx, body); err != nil {
				wait := time.Duration(attempt) * retryBaseWait
				log.Warn().
					Err(err).
					Int("attempt", attempt).
					Dur("retry_in", wait).
					Msg("webhook: post failed, retrying")

				if attempt < maxRetries {
					select {
					case <-ctx.Done():
						return
					case <-time.After(wait):
					}
				}
				continue
			}
			log.Debug().Int("attempt", attempt).Msg("webhook: event delivered")
			return
		}
		log.Error().Msg("webhook: all retry attempts exhausted")
	}()
}

// post sends the HTTP POST request with HMAC-SHA256 signature.
func (s *Sender) post(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Devert-Monitor-Agent/1.0")

	if s.secret != "" {
		sig := s.sign(body)
		req.Header.Set("X-Webhook-Signature", "sha256="+sig)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// sign computes an HMAC-SHA256 hex digest of body using the configured secret.
func (s *Sender) sign(body []byte) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
