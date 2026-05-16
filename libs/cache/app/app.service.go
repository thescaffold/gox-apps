package app

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/thescaffold/gox-apps/libs/cache/app/list"
)

// AppService backs the cache endpoints. Mirrors ntx-apps/libs/cache/src/app.service.ts
// plus the additional helpers (setnx/getset/ttl/incr) that TS keeps directly
// on the controller.
type AppService struct {
	listService *list.ListService `inject:""`
}

// GetHello mirrors TS app.service.ts AppService.getHello().
func (s *AppService) GetHello() string { return "Hello World!" }

// Get retrieves the entry for key, applying the (buggy) TS expiry filter.
func (s *AppService) Get(key string) (*list.List, error) {
	return s.listService.Get(key)
}

// Set upserts key→value with duration in MILLISECONDS — TS passes `duration`
// to dayjs `.add(duration)` which defaults to the millisecond unit.
func (s *AppService) Set(key string, value any, durationMs int64) (*list.List, error) {
	return s.listService.Set(key, value, asDuration(durationMs))
}

// SetNx attempts to insert the entry only when key does not exist.
// Returns 1 on insert, 0 when the key was already present. Mirrors TS
// setNx() which uses INSERT … orIgnore() and reads result.identifiers.length.
func (s *AppService) SetNx(key string, value any, durationMs int64) (int, error) {
	existing, _ := s.listService.GetAny(key)
	if existing != nil {
		return 0, nil
	}
	if _, err := s.listService.Set(key, value, asDuration(durationMs)); err != nil {
		return 0, err
	}
	return 1, nil
}

// GetSet atomically retrieves the previous entry then overwrites it. Mirrors
// TS getSet() (transaction + pessimistic_write lock); atomicity is provided
// by ListService.GetSet's mutex.
func (s *AppService) GetSet(key string, value any, durationMs int64) (*list.List, error) {
	return s.listService.GetSet(key, value, asDuration(durationMs))
}

// TTL returns remaining lifetime in seconds for key.
//
//	-2: key not found OR expired
//	-1: key has no expiry
//	>=0: seconds until expiry
//
// Mirrors TS app.controller.ts ttl(): floor((expiredAt - now) / 1000), and
// translates any negative seconds back to -2.
func (s *AppService) TTL(key string) (int64, error) {
	record, _ := s.listService.GetAny(key)
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

// Incr atomically increments the integer value at key. A missing key starts at
// 1; a non-numeric value yields 0 (matching TS NaN behaviour). TS clears the
// expiry on every Incr (`expiredAt: null`), which this implementation mirrors
// by passing a zero duration through.
func (s *AppService) Incr(key string) (int64, error) {
	record, _ := s.listService.GetAny(key)
	if record == nil {
		if _, err := s.listService.Set(key, 1, 0); err != nil {
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
	if _, err := s.listService.Set(key, current, 0); err != nil {
		return 0, err
	}
	return current, nil
}

// Del removes the entry at key and returns the number of rows affected
// (0 when nothing matched). Mirrors TS `deleteResult.affected || 0`.
func (s *AppService) Del(key string) (int64, error) {
	return s.listService.Del(key)
}

// asDuration converts a millisecond count from the wire DTO into time.Duration.
// A zero or negative input yields zero (i.e. no expiry), matching TS where a
// falsy `duration` short-circuits to `expiredAt: null`.
func asDuration(ms int64) time.Duration {
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
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
