package channel

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// ChannelController mirrors ntx-apps/libs/polylog/src/api/channel/channel.controller.ts.
type ChannelController struct {
	crud.CrudResource[Channel, CreateChannelDto, UpdateChannelDto]

	entity *ChannelEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *ChannelController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Channel, CreateChannelDto, UpdateChannelDto]{
		// Mirrors TS channel.controller.ts:19 name = 'channel'.
		Name: "channel",
		// Mirrors TS channel.controller.ts:20 searchable = ['tags'].
		Searchable: []string{"tags"},
		// Mirrors TS channel.controller.ts:22-24 unique = ({workspaceId,sinkId,type})
		// => [{workspaceId,sinkId,type}].
		Unique: func(d *CreateChannelDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"sink_id":      d.SinkId,
				"type":         d.Type,
			}}
		},
		// Mirrors TS channel.controller.ts:26-37 morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateChannelDto)
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
