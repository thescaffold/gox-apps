package device

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// DeviceController mirrors ntx-apps/libs/identity/src/api/device/device.controller.ts.
type DeviceController struct {
	crud.CrudResource[Device, CreateDeviceDto, UpdateDeviceDto]

	entity *DeviceEntity `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *DeviceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Device, CreateDeviceDto, UpdateDeviceDto]{
		// Mirrors TS device.controller.ts name = 'device'.
		Name: "device",
		// Mirrors TS device.controller.ts searchable = ['os','agent','engine','cpu'].
		// gox Device entity has fingerprint/type/platform instead of os/agent/
		// engine/cpu; the TS list is kept verbatim for API surface parity.
		Searchable: []string{"os", "agent", "engine", "cpu"},
		// TS unique = () => [] — no uniqueness constraints; left unset.
	})
	c.SetLang(c.lang)
}
