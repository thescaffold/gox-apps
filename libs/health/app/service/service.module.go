package service

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ServiceModule struct{}

func (m *ServiceModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ServiceModule) Exports() []any { return []any{&ServiceService{}, &ServiceEntity{}} }

func (m *ServiceModule) Declarations() []any {
	return []any{&ServiceService{}, &ServiceEntity{}}
}
