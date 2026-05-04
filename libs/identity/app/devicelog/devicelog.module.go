package devicelog

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type DeviceLogModule struct{}

func (m *DeviceLogModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *DeviceLogModule) Exports() []any { return []any{&DeviceLogService{}} }

func (m *DeviceLogModule) Declarations() []any {
	return []any{&DeviceLogService{}, &DeviceLogEntity{}}
}
