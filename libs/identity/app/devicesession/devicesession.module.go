package devicesession

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type DeviceSessionModule struct{}

func (m *DeviceSessionModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *DeviceSessionModule) Exports() []any { return []any{&DeviceSessionService{}} }

func (m *DeviceSessionModule) Declarations() []any {
	return []any{&DeviceSessionService{}, &DeviceSessionEntity{}}
}
