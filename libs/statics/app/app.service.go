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

// FilterByCode mirrors TS AppService.filterByCode() — same behaviour as
// FindByCode, exposed under the TS spelling so cross-package callers
// (bridge.LicenseType, capital.PlanType) can use either name interchangeably.
func (s *AppService) FilterByCode(code string, parentId *string) (*list.List, error) {
	return s.listService.FindByCode(code, parentId)
}

// FilterByValue mirrors TS AppService.filterByValue().
func (s *AppService) FilterByValue(value string, parentId *string) (*list.List, error) {
	return s.listService.FindByValue(value, parentId)
}
