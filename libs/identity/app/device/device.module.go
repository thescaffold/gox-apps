package device

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type DeviceModule struct{}

func (m *DeviceModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *DeviceModule) Exports() []any { return []any{&DeviceService{}} }

func (m *DeviceModule) Declarations() []any {
	return []any{&DeviceService{}, &DeviceEntity{}}
}
