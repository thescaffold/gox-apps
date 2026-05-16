package log

import (
	"errors"
	"os"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	queueapp "github.com/thescaffold/gox-apps/libs/queue/app"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// LogController mirrors ntx-apps/libs/notification/src/api/log/log.controller.ts.
// Phase 7.3 wires the BeforeCreate passCode gate and the AfterCreate queue
// push to `queue/apps/notification/message` (the consumer drives the
// notification provider — see notification/app/index.go for the handler).
type LogController struct {
	crud.CrudResource[Log, CreateLogDto, UpdateLogDto]

	entity *LogEntity         `inject:""`
	queue  *queueapp.AppService `inject:""`
	lang   *i18n.Service      `inject:""`
}

func (c *LogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Log, CreateLogDto, UpdateLogDto]{
		// Mirrors TS log.controller.ts:27 name = 'log'.
		Name: "log",
		// Mirrors TS log.controller.ts:28 searchable = ['subject','message'].
		Searchable: []string{"subject", "message"},
		// Mirrors TS log.controller.ts:30 unique = ({reference}) => [{reference}].
		Unique: func(d *CreateLogDto) []map[string]any {
			var ref any
			if d.Reference != nil {
				ref = *d.Reference
			}
			return []map[string]any{{"reference": ref}}
		},
		// Mirrors TS log.controller.ts:33-65 morphs.beforeCreate: spreads
		// userId/clientId/workspaceId from context, defaults publishAt → now
		// and expireAt → publishAt+3day when missing.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateLogDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = &id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = &id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = &id
					}
				}
				if p.PublishAt == nil {
					now := time.Now().UTC()
					p.PublishAt = &now
				}
				if p.ExpireAt == nil {
					exp := p.PublishAt.Add(3 * 24 * time.Hour)
					p.ExpireAt = &exp
				}
				return p, nil
			},
		},
		// Mirrors TS log.controller.ts:67-119:
		//   - BeforeCreate hook: passCode gate when ctx.user is absent — block
		//     unauthenticated creates that don't supply the correct passCode.
		//   - AfterCreate hook: enqueue the saved Log for ProviderService.send
		//     processing on "queue/apps/notification" (job: "message").
		Hooks: map[string]crud.HookFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) error {
				p, ok := payload.(*CreateLogDto)
				if !ok || p == nil {
					return nil
				}
				// When the request carries no user, require the static PASS_CODE
				// secret to match. Matches TS behaviour: any request that comes
				// through the public bus must prove it's authorised.
				if ctx.User == nil {
					expected := os.Getenv("PASS_CODE")
					supplied := ""
					if p.Passcode != nil {
						supplied = *p.Passcode
					}
					if expected == "" || supplied != expected {
						return errors.New("not allowed")
					}
				}
				return nil
			},
			crud.AfterCreate: func(payload any, _ ntxctx.NTXContext) error {
				if c.queue == nil {
					return nil
				}
				row, ok := payload.(*Log)
				if !ok || row == nil {
					return nil
				}
				ref := ""
				if row.Reference != "" {
					ref = row.Reference
				}
				// TS retryLimit:3, retryDelay:60 (seconds → ms in goose).
				cfg := &goqueues.JobConfig{
					RetryLimit: 3,
					RetryDelay: 60 * 1000,
				}
				_, _ = c.queue.Push("queue/apps/notification", "message", map[string]any{
					"log":       row,
					"key":       row.Key,
					"reference": ref,
				}, cfg)
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}
