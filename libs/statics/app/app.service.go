package app

import (
	"github.com/thescaffold/gox-apps-statics/app/list"
)

type AppService struct {
	listService *list.ListService `inject:""`
}

func (s *AppService) GetHello() string {
	return "Hello from statics"
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
