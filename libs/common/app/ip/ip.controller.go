package ip

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// IpController mirrors ntx-apps/libs/common/src/api/ip/ip.controller.ts.
type IpController struct {
	crud.CrudResource[Ip, CreateIpDto, UpdateIpDto]

	entity *IpEntity     `inject:""`
	lang   *i18n.Service `inject:""`
}

func (c *IpController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Ip, CreateIpDto, UpdateIpDto]{
		// Mirrors TS ip.controller.ts:18 name = 'ip'.
		Name: "ip",
		// Mirrors TS ip.controller.ts:19 searchable = ['value','currency','city']
		// — `currency` is a TS placeholder; the Ip entity has no `currency`
		// column so this field never participates in search, but we preserve
		// the surface identically.
		Searchable: []string{"value", "currency", "city"},
		// Mirrors TS ip.controller.ts:21 unique = ({country}) => [{country}].
		Unique: func(d *CreateIpDto) []map[string]any {
			return []map[string]any{{"country": d.Country}}
		},
	})
	c.SetLang(c.lang)
}
