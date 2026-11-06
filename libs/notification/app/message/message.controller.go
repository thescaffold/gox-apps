package message

import (
	"fmt"
	"sync/atomic"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// MessageController mirrors ntx-apps/libs/notification/src/api/message/message.controller.ts.
var ownerScoped atomic.Bool

// ScopeToOwner confines the message API to the caller's own messages in their
// workspace (crud WorkspaceScoped + OwnerScoped): for a host where messages are
// a person's inbox. Without it, as before, any signed-in caller can list every
// message. Call it before the kernel boots.
func ScopeToOwner() { ownerScoped.Store(true) }

type MessageController struct {
	crud.CrudResource[Message, CreateMessageDto, UpdateMessageDto]

	entity *MessageEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *MessageController) OnRegister() {
	c.Hydrate(c.entity, config())
	c.SetLang(c.lang)
}

// config is the resource's crud configuration, read at registration: with
// ScopeToOwner called first, every route is confined to the caller's own
// messages in their workspace and fails closed without a verified caller.
func config() crud.Config[Message, CreateMessageDto, UpdateMessageDto] {
	scoped := ownerScoped.Load()
	return crud.Config[Message, CreateMessageDto, UpdateMessageDto]{
		WorkspaceScoped: scoped,
		OwnerScoped:     scoped,
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
				// A caller verified from token claims carries only the ids.
				if p.UserId == nil && ctx.UserID != "" {
					id := ctx.UserID
					p.UserId = &id
				}
				if p.ClientId == nil && ctx.ClientID != "" {
					id := ctx.ClientID
					p.ClientId = &id
				}
				if p.WorkspaceId == nil && ctx.WorkspaceID != "" {
					id := ctx.WorkspaceID
					p.WorkspaceId = &id
				}
				// Scoped to the owner, a message always belongs to its creator:
				// nobody files a message into someone else's inbox through it.
				if scoped {
					if ctx.UserID == "" || ctx.WorkspaceID == "" {
						return nil, fmt.Errorf("message: a verified caller is required")
					}
					uid, ws := ctx.UserID, ctx.WorkspaceID
					p.UserId, p.WorkspaceId = &uid, &ws
				}
				return p, nil
			},
		},
	}
}
