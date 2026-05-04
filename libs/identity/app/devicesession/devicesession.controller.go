package devicesession

import "github.com/thescaffold/gox-packages-core/crud"

type DeviceSessionController struct {
	crud.CrudResource[DeviceSession, CreateDeviceSessionDto, UpdateDeviceSessionDto]
	entity *DeviceSessionEntity `inject:""`
}

func (c *DeviceSessionController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[DeviceSession, CreateDeviceSessionDto, UpdateDeviceSessionDto]{
		Name:       "IdentityDeviceSession",
		Searchable: []string{"device_id", "user_id"},
	})
}
