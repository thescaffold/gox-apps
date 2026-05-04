package list

import (
	"encoding/json"
	"time"
)

type ListService struct {
	entity *ListEntity `inject:""`
}

func (s *ListService) Get(key, group string) (*List, error) {
	now := time.Now().UTC()
	return s.entity.First("key = ? AND \"group\" = ? AND (expired_at IS NULL OR expired_at > ?)", key, group, now)
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
