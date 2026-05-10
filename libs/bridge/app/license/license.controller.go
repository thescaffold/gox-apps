package license

import "github.com/thescaffold/gox-packages/libs/core/crud"

type LicenseController struct {
	crud.CrudResource[License, CreateLicenseDto, UpdateLicenseDto]
	entity *LicenseEntity `inject:""`
}

func (c *LicenseController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[License, CreateLicenseDto, UpdateLicenseDto]{
		Name:       "BridgeLicense",
		Searchable: []string{"user_id", "workspace_id", "type_id"},
	})
}
