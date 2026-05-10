package ip

import "github.com/thescaffold/gox-packages/libs/core/crud"

type IpController struct {
	crud.CrudResource[Ip, CreateIpDto, UpdateIpDto]
	entity *IpEntity `inject:""`
}

func (c *IpController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Ip, CreateIpDto, UpdateIpDto]{
		Name:       "CommonIp",
		Searchable: []string{"value", "country", "city"},
	})
}
