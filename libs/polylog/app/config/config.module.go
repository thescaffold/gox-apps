package config

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ConfigModule struct{}

func (m *ConfigModule) Imports() []types.Module {
	return []types.Module{sql.Child(&sql.Config{})}
}

func (m *ConfigModule) Exports() []any { return []any{&ConfigService{}} }

func (m *ConfigModule) Declarations() []any {
	return []any{&ConfigService{}, &ConfigEntity{}}
}
