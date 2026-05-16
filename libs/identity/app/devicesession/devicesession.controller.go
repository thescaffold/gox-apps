package devicesession

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// DeviceSessionController mirrors ntx-apps/libs/identity/src/api/device-session/device-session.controller.ts.
type DeviceSessionController struct {
	crud.CrudResource[DeviceSession, CreateDeviceSessionDto, UpdateDeviceSessionDto]

	entity *DeviceSessionEntity `inject:""`
	lang   *i18n.Service        `inject:""`
}

func (c *DeviceSessionController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[DeviceSession, CreateDeviceSessionDto, UpdateDeviceSessionDto]{
		// Mirrors TS device-session.controller.ts name = 'device session'.
		Name: "device session",
		// Mirrors TS device-session.controller.ts searchable = [] (empty).
		Searchable: []string{},
		// Mirrors TS device-session.controller.ts unique = ({deviceId}) =>
		// [{deviceId}].
		Unique: func(d *CreateDeviceSessionDto) []map[string]any {
			return []map[string]any{{"device_id": d.DeviceId}}
		},
		// Mirrors TS device-session.controller.ts morphs.beforeCreate — TS only
		// spreads userId from context.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateDeviceSessionDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}
