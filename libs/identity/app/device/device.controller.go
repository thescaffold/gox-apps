package device

import "github.com/thescaffold/gox-packages/libs/core/crud"

type DeviceController struct {
	crud.CrudResource[Device, CreateDeviceDto, UpdateDeviceDto]
	entity *DeviceEntity `inject:""`
}

func (c *DeviceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Device, CreateDeviceDto, UpdateDeviceDto]{
		Name:       "IdentityDevice",
		Searchable: []string{"user_id", "fingerprint", "type"},
	})
}
