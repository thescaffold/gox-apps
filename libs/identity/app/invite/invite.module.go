package invite

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type InviteModule struct{}

func (m *InviteModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *InviteModule) Exports() []any { return []any{&InviteService{}} }

func (m *InviteModule) Declarations() []any {
	return []any{&InviteService{}, &InviteEntity{}}
}
