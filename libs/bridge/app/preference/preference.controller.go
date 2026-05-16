package preference

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// PreferenceController mirrors ntx-apps/libs/bridge/src/api/preference/preference.controller.ts.
type PreferenceController struct {
	crud.CrudResource[Preference, CreatePreferenceDto, UpdatePreferenceDto]

	entity *PreferenceEntity `inject:""`
	lang   *i18n.Service     `inject:""`
}

func (c *PreferenceController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Preference, CreatePreferenceDto, UpdatePreferenceDto]{
		// Mirrors TS preference.controller.ts:18 name = 'preference'.
		Name: "preference",
		// Mirrors TS preference.controller.ts:19 searchable = ['value'].
		Searchable: []string{"value"},
		// Mirrors TS preference.controller.ts:21-23 unique = ({licenseId,type,
		// key}) => [{licenseId,type,key}].
		Unique: func(d *CreatePreferenceDto) []map[string]any {
			return []map[string]any{{
				"license_id": d.LicenseId,
				"type":       d.Type,
				"key":        d.Key,
			}}
		},
	})
	c.SetLang(c.lang)
}
