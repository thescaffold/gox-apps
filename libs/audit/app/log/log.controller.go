package log

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/audit/pkg/record"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// LogController mirrors ntx-apps/libs/audit/src/api/log/log.controller.ts.
type LogController struct {
	crud.CrudResource[Log, CreateLogDto, UpdateLogDto]

	entity *LogEntity    `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *LogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Log, CreateLogDto, UpdateLogDto]{
		// Mirrors TS log.controller.ts:19 name = 'log'.
		Name: "log",
		// Mirrors TS log.controller.ts:20 searchable = ['desc'].
		Searchable: []string{"desc"},
		// TRD U-S11: Log.WorkspaceId is NOT NULL (no global rows) — an
		// audit trail is exactly the kind of data a cross-tenant read
		// would be most damaging for.
		WorkspaceScoped: true,
		// Mirrors TS morphs.beforeCreate which spreads
		// `{ userId: user.id, clientId: client.id, workspaceId: workspace.id }`
		// onto the create payload. The morph runs before the entity insert so
		// the values flow through copyAny onto the Log row's NOT NULL columns.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateLogDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok {
						p.WorkspaceId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}

// With append-only on (record.EnableAppendOnly), entries cannot be changed or
// removed through the API. Without it these behave exactly as before.

func appendOnlyRefusal() types.Output {
	return response.Forbidden("audit", "the audit log is append-only")
}

func (c *LogController) Update(dto *crud.UpdateDto[UpdateLogDto]) types.Output {
	if record.AppendOnly() {
		return appendOnlyRefusal()
	}
	return c.CrudResource.Update(dto)
}

func (c *LogController) UpdatePut(dto *crud.UpdatePutDto[CreateLogDto]) types.Output {
	if record.AppendOnly() {
		return appendOnlyRefusal()
	}
	return c.CrudResource.UpdatePut(dto)
}

func (c *LogController) Upsert(dto *crud.UpsertDto[CreateLogDto]) types.Output {
	if record.AppendOnly() {
		return appendOnlyRefusal()
	}
	return c.CrudResource.Upsert(dto)
}

func (c *LogController) Delete(dto *crud.DeleteDto) types.Output {
	if record.AppendOnly() {
		return appendOnlyRefusal()
	}
	return c.CrudResource.Delete(dto)
}
