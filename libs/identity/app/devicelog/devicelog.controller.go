package devicelog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// DeviceLogController mirrors ntx-apps/libs/identity/src/api/device-log/device-log.controller.ts.
type DeviceLogController struct {
	crud.CrudResource[DeviceLog, CreateDeviceLogDto, UpdateDeviceLogDto]

	entity *DeviceLogEntity `inject:""`
	lang   *i18n.Service    `inject:""`
}

func (c *DeviceLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[DeviceLog, CreateDeviceLogDto, UpdateDeviceLogDto]{
		// Mirrors TS device-log.controller.ts name = 'devicelog' (one word — TS
		// uses a typo'd value, kept verbatim for surface parity).
		Name: "devicelog",
		// Mirrors TS device-log.controller.ts searchable = ['ip','country','region','city','area'].
		// gox DeviceLog entity has device_id/event/meta instead; the TS list is
		// kept verbatim for API surface parity.
		Searchable: []string{"ip", "country", "region", "city", "area"},
		// TS unique = () => [] — no uniqueness constraints; left unset.
	})
	c.SetLang(c.lang)
}
