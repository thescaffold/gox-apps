package userclientworkspace

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type UserClientWorkspaceModule struct{}

func (m *UserClientWorkspaceModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *UserClientWorkspaceModule) Exports() []any { return []any{&UserClientWorkspaceService{}} }

func (m *UserClientWorkspaceModule) Declarations() []any {
	return []any{&UserClientWorkspaceService{}, &UserClientWorkspaceEntity{}}
}
