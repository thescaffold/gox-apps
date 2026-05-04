package licensetype

import "github.com/thescaffold/gox-packages-core/crud"

type LicenseTypeController struct {
	crud.CrudResource[LicenseType, CreateLicenseTypeDto, UpdateLicenseTypeDto]
	entity *LicenseTypeEntity `inject:""`
}

func (c *LicenseTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[LicenseType, CreateLicenseTypeDto, UpdateLicenseTypeDto]{
		Name:       "BridgeLicenseType",
		Searchable: []string{"key", "name"},
	})
}
