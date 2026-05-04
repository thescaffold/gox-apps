package job

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type JobModule struct{}

func (m *JobModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *JobModule) Exports() []any { return []any{&JobService{}} }

func (m *JobModule) Declarations() []any {
	return []any{&JobService{}, &JobEntity{}}
}
