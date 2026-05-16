package list

import (
	"encoding/json"
	"sync"
	"time"
)

// ListService backs the cache app endpoints. Operations are keyed solely on
// `key` — the database's `group` column is defaulted to '' by the migration,
// matching TS where the entity has no group concept.
type ListService struct {
	entity *ListEntity `inject:""`
	// getSetMu serialises GetSet so the read-then-write pair is atomic per
	// process — mirrors TS's pessimistic_write transaction lock.
	getSetMu sync.Mutex
}

// Get mirrors TS app.service.ts get() / app.controller.ts get() exactly,
// including its bug: TypeORM's `where: [{key, expiredAt:IsNull()}, {key, expiredAt:LessThan(now)}]`
// is an OR, so it returns rows with no expiry OR rows that are ALREADY
// expired — and never returns an unexpired row that has an expiry set. We
// match the bug verbatim so observable behaviour stays identical; flip the
// `<` to `>` if you ever fix the TS side.
func (s *ListService) Get(key string) (*List, error) {
	now := time.Now().UTC()
	return s.entity.First(`key = ? AND (expired_at IS NULL OR expired_at < ?)`, key, now)
}

// GetAny returns the entry for key regardless of expiry. Mirrors TS
// findOne({ where: { key } }) without the LessThan filter — used by ttl()
// and incr() which inspect expired records.
func (s *ListService) GetAny(key string) (*List, error) {
	return s.entity.First("key = ?", key)
}

// GetSet atomically reads the existing entry then overwrites it with value.
// Returns the *previous* entry (or nil when none existed), matching TS getSet().
func (s *ListService) GetSet(key string, value any, ttl time.Duration) (*List, error) {
	s.getSetMu.Lock()
	defer s.getSetMu.Unlock()

	prev, _ := s.entity.First("key = ?", key)

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var expiredAt *time.Time
	if ttl > 0 {
		t := time.Now().UTC().Add(ttl)
		expiredAt = &t
	}

	if prev != nil {
		// Copy then update to keep prev unchanged for the caller.
		updated := *prev
		updated.Value = valueJSON
		updated.ExpiredAt = expiredAt
		if _, err := s.entity.Update(&updated, "id = ?", prev.Id); err != nil {
			return nil, err
		}
		return prev, nil
	}

	item := &List{Key: key, Value: valueJSON, ExpiredAt: expiredAt}
	if err := s.entity.Insert(item); err != nil {
		return nil, err
	}
	return nil, nil
}

// Set upserts the entry at key. Mirrors TS updateOrCreate({key}, {value, expiredAt}).
func (s *ListService) Set(key string, value any, ttl time.Duration) (*List, error) {
	var expiredAt *time.Time
	if ttl > 0 {
		t := time.Now().UTC().Add(ttl)
		expiredAt = &t
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	existing, _ := s.entity.First("key = ?", key)
	if existing != nil {
		existing.Value = valueJSON
		existing.ExpiredAt = expiredAt
		if _, err := s.entity.Update(existing, "id = ?", existing.Id); err != nil {
			return nil, err
		}
		return existing, nil
	}

	item := &List{Key: key, Value: valueJSON, ExpiredAt: expiredAt}
	if err := s.entity.Insert(item); err != nil {
		return nil, err
	}
	return item, nil
}

// Del removes the entry at key. Mirrors TS listService.repository.delete({key}).
// Returns the number of rows deleted so the controller can echo the count
// (TS returns `deleteResult.affected || 0`).
func (s *ListService) Del(key string) (int64, error) {
	existing, err := s.entity.First("key = ?", key)
	if err != nil || existing == nil {
		return 0, nil
	}
	n, err := s.entity.Delete("id = ?", existing.Id)
	if err != nil {
		return 0, err
	}
	return n, nil
}
