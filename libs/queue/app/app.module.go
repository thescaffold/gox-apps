package app

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/types"
	queuejob "github.com/thescaffold/gox-apps/libs/queue/app/job"
	queuelog "github.com/thescaffold/gox-apps/libs/queue/app/log"
	queuequeue "github.com/thescaffold/gox-apps/libs/queue/app/queue"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		goqueues.NewModule(nil, nil, true),
		&queuequeue.QueueModule{},
		&queuejob.JobModule{},
		&queuelog.LogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	return []any{&AppController{}, &AppService{}}
}
