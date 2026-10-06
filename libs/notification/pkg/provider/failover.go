package provider

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Provider failover. A channel's provider setting may list several providers,
// "termii,africastalking": they are tried in order, a provider that fails
// several times in a row is skipped for a cooldown, and the first that accepts
// the message wins. A single name behaves exactly as before.

var (
	customMu    sync.RWMutex
	customEmail = map[string]func(EmailMessage) bool{}
	customSMS   = map[string]func(SMSMessage) bool{}
)

// RegisterEmailProvider adds (or replaces) a named email provider.
func RegisterEmailProvider(name string, send func(EmailMessage) bool) {
	customMu.Lock()
	customEmail[name] = send
	customMu.Unlock()
}

// RegisterSMSProvider adds (or replaces) a named SMS provider.
func RegisterSMSProvider(name string, send func(SMSMessage) bool) {
	customMu.Lock()
	customSMS[name] = send
	customMu.Unlock()
}

func customEmailFor(name string) (func(EmailMessage) bool, bool) {
	customMu.RLock()
	defer customMu.RUnlock()
	f, ok := customEmail[name]
	return f, ok
}

func customSMSFor(name string) (func(SMSMessage) bool, bool) {
	customMu.RLock()
	defer customMu.RUnlock()
	f, ok := customSMS[name]
	return f, ok
}

// splitProviders turns "a, b ,c" into [a b c], dropping empties.
func splitProviders(list string) []string {
	var out []string
	for _, p := range strings.Split(list, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

type circuit struct {
	fails     int
	openUntil time.Time
}

var (
	circuitMu sync.Mutex
	circuits  = map[string]*circuit{}
	nowFn     = time.Now
)

func breakerSettings() (threshold int, cooldown time.Duration) {
	threshold, cooldown = 3, 60*time.Second
	if n, err := strconv.Atoi(os.Getenv("NOTIFICATION_BREAKER_FAILS")); err == nil && n > 0 {
		threshold = n
	}
	if n, err := strconv.Atoi(os.Getenv("NOTIFICATION_BREAKER_COOLDOWN_SECS")); err == nil && n > 0 {
		cooldown = time.Duration(n) * time.Second
	}
	return
}

// circuitOpen reports whether a provider should be skipped right now. When the
// cooldown has passed it lets one attempt through (half-open).
func circuitOpen(channel, name string) bool {
	threshold, _ := breakerSettings()
	circuitMu.Lock()
	defer circuitMu.Unlock()
	c := circuits[channel+"/"+name]
	if c == nil || c.fails < threshold {
		return false
	}
	if nowFn().Before(c.openUntil) {
		return true
	}
	c.openUntil = nowFn().Add(time.Second) // one trial; a failure re-opens below
	return false
}

func recordResult(channel, name string, ok bool) {
	threshold, cooldown := breakerSettings()
	circuitMu.Lock()
	defer circuitMu.Unlock()
	key := channel + "/" + name
	c := circuits[key]
	if c == nil {
		c = &circuit{}
		circuits[key] = c
	}
	if ok {
		c.fails = 0
		return
	}
	c.fails++
	if c.fails >= threshold {
		c.openUntil = nowFn().Add(cooldown)
	}
}

// resetCircuits clears breaker state (tests).
func resetCircuits() {
	circuitMu.Lock()
	circuits = map[string]*circuit{}
	circuitMu.Unlock()
}

// sendWithFailover tries each listed provider in order. When every circuit is
// open it still tries the first one rather than dropping the message.
func sendWithFailover(channel, list string, send func(name string) bool) bool {
	names := splitProviders(list)
	attempted := false
	for _, name := range names {
		if circuitOpen(channel, name) {
			continue
		}
		attempted = true
		ok := send(name)
		recordResult(channel, name, ok)
		if ok {
			return true
		}
	}
	if !attempted && len(names) > 0 {
		return send(names[0])
	}
	return false
}
