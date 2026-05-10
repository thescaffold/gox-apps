package preference

import "github.com/thescaffold/gox-packages/libs/core/crud"

type PreferenceController struct {
	crud.CrudResource[Preference, CreatePreferenceDto, UpdatePreferenceDto]
	entity *PreferenceEntity `inject:""`
}

func (c *PreferenceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Preference, CreatePreferenceDto, UpdatePreferenceDto]{
		Name: "BridgePreference", Searchable: []string{"license_id", "key"},
	})
}
