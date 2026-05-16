package message

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// MessageController mirrors ntx-apps/libs/notification/src/api/message/message.controller.ts.
type MessageController struct {
	crud.CrudResource[Message, CreateMessageDto, UpdateMessageDto]

	entity *MessageEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *MessageController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Message, CreateMessageDto, UpdateMessageDto]{
		// Mirrors TS message.controller.ts:20 name = 'message'.
		Name: "message",
		// Mirrors TS message.controller.ts:21 searchable = ['subject','message'].
		Searchable: []string{"subject", "message"},
		// Mirrors TS message.controller.ts:23 unique = ({reference}) => [{reference}].
		Unique: func(d *CreateMessageDto) []map[string]any {
			var ref any
			if d.Reference != nil {
				ref = *d.Reference
			}
			return []map[string]any{{"reference": ref}}
		},
		// Mirrors TS message.controller.ts:25-35 morphs.beforeCreate: spreads
		// userId/clientId/workspaceId from context.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateMessageDto)
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
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
