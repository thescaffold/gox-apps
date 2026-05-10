package devicelog

import "github.com/thescaffold/gox-packages/libs/core/crud"

type DeviceLogController struct {
	crud.CrudResource[DeviceLog, CreateDeviceLogDto, UpdateDeviceLogDto]
	entity *DeviceLogEntity `inject:""`
}

func (c *DeviceLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[DeviceLog, CreateDeviceLogDto, UpdateDeviceLogDto]{
		Name:       "IdentityDeviceLog",
		Searchable: []string{"device_id", "event"},
	})
}
