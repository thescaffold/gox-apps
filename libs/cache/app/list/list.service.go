package list

import (
	"encoding/json"
	"sync"
	"time"
)

type ListService struct {
	entity *ListEntity `inject:""`
	// getSetMu serialises GetSet so the read-then-write pair is atomic per
	// process — mirrors TS's pessimistic_write transaction lock.
	getSetMu sync.Mutex
}

func (s *ListService) Get(key, group string) (*List, error) {
	now := time.Now().UTC()
	return s.entity.First("key = ? AND \"group\" = ? AND (expired_at IS NULL OR expired_at > ?)", key, group, now)
}

// GetAny returns the entry for (key, group) regardless of expiry.
// Mirrors TS findOne({ where: { key } }) without LessThan filter.
func (s *ListService) GetAny(key, group string) (*List, error) {
	return s.entity.First("key = ? AND \"group\" = ?", key, group)
}

// GetSet atomically reads the existing entry then overwrites it with value.
// Returns the *previous* entry (or nil when none existed), matching TS getSet().
func (s *ListService) GetSet(key string, value any, group string, ttl time.Duration) (*List, error) {
	s.getSetMu.Lock()
	defer s.getSetMu.Unlock()

	prev, _ := s.entity.First("key = ? AND \"group\" = ?", key, group)

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

	item := &List{Key: key, Value: valueJSON, Group: group, ExpiredAt: expiredAt}
	if err := s.entity.Insert(item); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *ListService) Set(key string, value any, group string, ttl time.Duration) (*List, error) {
	var expiredAt *time.Time
	if ttl > 0 {
		t := time.Now().UTC().Add(ttl)
		expiredAt = &t
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	existing, _ := s.entity.First("key = ? AND \"group\" = ?", key, group)
	if existing != nil {
		existing.Value = valueJSON
		existing.ExpiredAt = expiredAt
		if _, err := s.entity.Update(existing, "id = ?", existing.Id); err != nil {
			return nil, err
		}
		return existing, nil
	}

	item := &List{Key: key, Value: valueJSON, Group: group, ExpiredAt: expiredAt}
	if err := s.entity.Insert(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ListService) Del(key, group string) error {
	existing, err := s.entity.First("key = ? AND \"group\" = ?", key, group)
	if err != nil || existing == nil {
		return nil
	}
	_, err = s.entity.Delete("id = ?", existing.Id)
	return err
}

func (s *ListService) Flush(group string) error {
	items, err := s.entity.Find(0, 0, "\"group\" = ?", group)
	if err != nil {
		return err
	}
	for _, item := range items {
		item := item
		if _, err := s.entity.Delete("id = ?", item.Id); err != nil {
			return err
		}
	}
	return nil
}
