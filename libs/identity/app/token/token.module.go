package token

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type TokenModule struct{}

func (m *TokenModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *TokenModule) Exports() []any { return []any{&TokenService{}} }

func (m *TokenModule) Declarations() []any {
	return []any{&TokenService{}, &TokenEntity{}}
}
