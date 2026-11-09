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
		// TranslationPaths mirror TS index.ts `paths = [translations/en/ntx/apps/${name}.yaml]`;
		// CoreModule.Boot pre-loads them so translate() resolves real strings.
		module.New(module.CoreConfig{
			TranslationPaths: Paths,
		}),
		goqueues.NewModule(nil, nil, true),
		&queuequeue.QueueModule{},
		&queuejob.JobModule{},
		&queuelog.LogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	// Until the container has run, queueAppSvc is this svc; AppService.OnRegister
	// then replaces it with the instance the container injected (svc itself
	// may never be: the container can fill this slot with the AppService it
	// already made for AppController's field — PLAN U18). A later call must
	// not put an uninjected one back.
	svc := &AppService{}
	if queueAppSvc == nil {
		queueAppSvc = svc
	}
	return []any{&AppController{}, svc}
}
