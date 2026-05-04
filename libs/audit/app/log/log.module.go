package log

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type LogModule struct{}

func (m *LogModule) Imports() []types.Module {
	return []types.Module{
		ROUTES,
		sql.Child(&sql.Config{}),
	}
}

func (m *LogModule) Exports() []any {
	return []any{&LogService{}}
}

func (m *LogModule) Declarations() []any {
	return []any{
		&LogService{},
		&LogEntity{},
	}
}
