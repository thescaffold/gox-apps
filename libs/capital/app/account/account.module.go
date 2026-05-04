package account

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type AccountModule struct{}

func (m *AccountModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *AccountModule) Exports() []any { return []any{&AccountService{}, &AccountEntity{}} }

func (m *AccountModule) Declarations() []any {
	return []any{&AccountService{}, &AccountEntity{}}
}
