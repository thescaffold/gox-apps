package app

import (
	"github.com/thescaffold/gox-apps/libs/statics/app/list"
)

type AppService struct {
	listService *list.ListService `inject:""`
}

// GetHello mirrors TS app.service.ts AppService.getHello().
func (s *AppService) GetHello() string {
	return "Hello World!"
}

func (s *AppService) FilterByKey(key string, parentId *string) ([]list.List, error) {
	return s.listService.FilterByKey(key, parentId)
}

func (s *AppService) FindByCode(code string, parentId *string) (*list.List, error) {
	return s.listService.FindByCode(code, parentId)
}

func (s *AppService) FindByValue(value string, parentId *string) (*list.List, error) {
	return s.listService.FindByValue(value, parentId)
}
