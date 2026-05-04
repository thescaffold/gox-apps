package app

import (
	"time"

	"github.com/thescaffold/gox-apps-cache/app/list"
)

type AppService struct {
	listService *list.ListService `inject:""`
}

func (s *AppService) GetHello() string { return "Hello from cache" }

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

func (s *AppService) Del(key, group string) error {
	return s.listService.Del(key, group)
}

func (s *AppService) Flush(group string) error {
	return s.listService.Flush(group)
}
