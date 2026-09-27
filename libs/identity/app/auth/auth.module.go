package auth

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type AuthModule struct{}

func (m *AuthModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *AuthModule) Exports() []any { return nil }

func (m *AuthModule) Declarations() []any {
	return []any{&AuthController{}}
}
