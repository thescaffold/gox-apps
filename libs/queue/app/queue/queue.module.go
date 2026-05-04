package queue

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type QueueModule struct{}

func (m *QueueModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *QueueModule) Exports() []any { return []any{&QueueService{}} }

func (m *QueueModule) Declarations() []any {
	return []any{&QueueService{}, &QueueEntity{}}
}
