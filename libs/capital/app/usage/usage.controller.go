package usage

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// UsageController mirrors ntx-apps/libs/capital/src/api/usage/usage.controller.ts.
type UsageController struct {
	crud.CrudResource[Usage, CreateUsageDto, UpdateUsageDto]

	entity *UsageEntity  `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *UsageController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Usage, CreateUsageDto, UpdateUsageDto]{
		// Mirrors TS usage.controller.ts name = 'usage'.
		Name: "usage",
		// Mirrors TS usage.controller.ts searchable = [] (empty).
		Searchable: []string{},
		// TS unique = () => [] — no uniqueness constraints; left unset.
		// Mirrors TS usage.controller.ts morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateUsageDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
