package workspace

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type WorkspaceModule struct{}

func (m *WorkspaceModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *WorkspaceModule) Exports() []any { return []any{&WorkspaceService{}} }

func (m *WorkspaceModule) Declarations() []any {
	return []any{&WorkspaceService{}, &WorkspaceEntity{}}
}
