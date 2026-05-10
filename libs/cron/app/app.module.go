package app

import (
	gocron "github.com/awesome-goose/goose/modules/cron"
	"github.com/awesome-goose/goose/types"
	cronjob "github.com/thescaffold/gox-apps-cron/app/job"
	cronlog "github.com/thescaffold/gox-apps-cron/app/log"
	"github.com/thescaffold/gox-packages-core/module"
)

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		gocron.NewModule(gocron.DefaultConfig(), Jobs, true),
		&cronjob.JobModule{},
		&cronlog.LogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	svc := &AppService{}
	cronAppSvc = svc
	return []any{&AppController{}, svc}
}
