package ip

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type IpModule struct{}

func (m *IpModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *IpModule) Exports() []any { return []any{&IpService{}} }

func (m *IpModule) Declarations() []any {
	return []any{&IpService{}, &IpEntity{}}
}
