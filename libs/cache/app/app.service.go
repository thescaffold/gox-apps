package app

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/thescaffold/gox-apps-cache/app/list"
)

type AppService struct {
	listService *list.ListService `inject:""`
}

// GetHello mirrors TS app.service.ts AppService.getHello().
func (s *AppService) GetHello() string { return "Hello World!" }

func (s *AppService) Get(key, group string) (*list.List, error) {
	return s.listService.Get(key, group)
}

func (s *AppService) Set(key string, value any, group string, ttlSecs int64) (*list.List, error) {
	var ttl time.Duration
	if ttlSecs > 0 {
		ttl = time.Duration(ttlSecs) * time.Second
	}
	return s.listService.Set(key, value, group, ttl)
}

// SetNx attempts to set the key only when it doesn't already exist.
// Returns 1 on insert, 0 when the key was already present.
// Mirrors TS app.controller.ts setNx() which uses INSERT … OR IGNORE.
func (s *AppService) SetNx(key string, value any, group string, ttlSecs int64) (int, error) {
	existing, _ := s.listService.GetAny(key, group)
	if existing != nil {
		return 0, nil
	}
	var ttl time.Duration
	if ttlSecs > 0 {
		ttl = time.Duration(ttlSecs) * time.Second
	}
	if _, err := s.listService.Set(key, value, group, ttl); err != nil {
		return 0, err
	}
	return 1, nil
}

// GetSet atomically retrieves the previous list entry then overwrites it.
// Mirrors TS app.controller.ts getSet() (transaction + pessimistic lock).
// The atomicity guarantee is provided by ListService.GetSet's mutex.
func (s *AppService) GetSet(key string, value any, group string, ttlSecs int64) (*list.List, error) {
	var ttl time.Duration
	if ttlSecs > 0 {
		ttl = time.Duration(ttlSecs) * time.Second
	}
	return s.listService.GetSet(key, value, group, ttl)
}

// TTL returns remaining lifetime in seconds for key.
//
//	-2: key not found
//	-1: key has no expiry
//	>=0: seconds until expiry
//
// Mirrors TS app.controller.ts ttl().
func (s *AppService) TTL(key, group string) (int64, error) {
	record, _ := s.listService.GetAny(key, group)
	if record == nil {
		return -2, nil
	}
	if record.ExpiredAt == nil {
		return -1, nil
	}
	seconds := int64(math.Floor(time.Until(*record.ExpiredAt).Seconds()))
	if seconds < 0 {
		return -2, nil
	}
	return seconds, nil
}

// Incr atomically increments the integer value at key.
// Missing key → 1. Non-numeric value → 0 (matches TS NaN behaviour).
// Mirrors TS app.controller.ts incr().
func (s *AppService) Incr(key, group string) (int64, error) {
	record, _ := s.listService.GetAny(key, group)
	if record == nil {
		if _, err := s.listService.Set(key, 1, group, 0); err != nil {
			return 0, err
		}
		return 1, nil
	}

	var raw any
	if len(record.Value) > 0 {
		_ = json.Unmarshal(record.Value, &raw)
	}
	current, ok := coerceInt(raw)
	if !ok {
		return 0, nil
	}
	current++
	if _, err := s.listService.Set(key, current, group, 0); err != nil {
		return 0, err
	}
	return current, nil
}

func coerceInt(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int:
		return int64(x), true
	case int64:
		return x, true
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func (s *AppService) Del(key, group string) error {
	return s.listService.Del(key, group)
}

func (s *AppService) Flush(group string) error {
	return s.listService.Flush(group)
}
