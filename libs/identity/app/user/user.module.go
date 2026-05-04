package user

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type UserModule struct{}

func (m *UserModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *UserModule) Exports() []any { return []any{&UserService{}} }

func (m *UserModule) Declarations() []any {
	return []any{&UserService{}, &UserEntity{}}
}
