package route

type CreateRouteDto struct {
	Group    string  `json:"group"    binding:"required"`
	Service  string  `json:"service"  binding:"required"`
	Type     string  `json:"type"`
	Name     string  `json:"name"     binding:"required"`
	Desc     *string `json:"desc,omitempty"`
	Upstream string  `json:"upstream" binding:"required"`
	Status   *string `json:"status,omitempty"`
}

type UpdateRouteDto struct {
	Group    *string `json:"group,omitempty"`
	Service  *string `json:"service,omitempty"`
	Type     *string `json:"type,omitempty"`
	Name     *string `json:"name,omitempty"`
	Desc     *string `json:"desc,omitempty"`
	Upstream *string `json:"upstream,omitempty"`
	Status   *string `json:"status,omitempty"`
}
