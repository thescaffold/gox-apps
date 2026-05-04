package task

import (
	"github.com/awesome-goose/goose/types"
	healthlog "github.com/thescaffold/gox-apps-health/app/log"
	healthsvc "github.com/thescaffold/gox-apps-health/app/service"
	healthsum "github.com/thescaffold/gox-apps-health/app/summary"
)

type TaskModule struct{}

func (m *TaskModule) Imports() []types.Module {
	return []types.Module{
		&healthlog.LogModule{},
		&healthsvc.ServiceModule{},
		&healthsum.SummaryModule{},
	}
}

func (m *TaskModule) Exports() []any { return nil }

func (m *TaskModule) Declarations() []any {
	return []any{&SummaryTask{}}
}
