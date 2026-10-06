package provider

import (
	"os"
	"strings"
	"sync"
	"time"
)

// The capture provider stores messages instead of sending them, so tests can
// read OTPs and links. It is selected with NOTIFICATION_DEFAULT_PROVIDER=capture
// (or a channel override) and is refused when the environment is production.

// Captured is one stored message.
type Captured struct {
	Channel string    `json:"channel"`
	To      string    `json:"to"`
	Subject string    `json:"subject"`
	Text    string    `json:"text"`
	HTML    string    `json:"html"`
	At      time.Time `json:"at"`
}

const captureLimit = 1000

var (
	captureMu  sync.Mutex
	captureLog []Captured
)

// IsProduction reports whether the process runs in production, where capture
// must never be available.
func IsProduction() bool {
	for _, k := range []string{"NODE_ENV", "APP_ENV"} {
		if strings.EqualFold(os.Getenv(k), "production") {
			return true
		}
	}
	return false
}

func capture(c Captured) bool {
	if IsProduction() {
		return false
	}
	c.At = time.Now()
	captureMu.Lock()
	defer captureMu.Unlock()
	captureLog = append(captureLog, c)
	if len(captureLog) > captureLimit {
		captureLog = captureLog[len(captureLog)-captureLimit:]
	}
	return true
}

// Captures returns captured messages, newest first, optionally only those sent
// to `to`, at most limit (default 50).
func Captures(to string, limit int) []Captured {
	if limit <= 0 {
		limit = 50
	}
	captureMu.Lock()
	defer captureMu.Unlock()
	out := []Captured{}
	for i := len(captureLog) - 1; i >= 0 && len(out) < limit; i-- {
		if to == "" || strings.EqualFold(captureLog[i].To, to) {
			out = append(out, captureLog[i])
		}
	}
	return out
}

// ClearCaptures empties the store (tests).
func ClearCaptures() {
	captureMu.Lock()
	captureLog = nil
	captureMu.Unlock()
}

// CaptureEnabled reports whether the read endpoint is available: never in
// production, and only when NOTIFICATION_CAPTURE_TOKEN is set.
func CaptureEnabled() bool {
	return !IsProduction() && os.Getenv("NOTIFICATION_CAPTURE_TOKEN") != ""
}
