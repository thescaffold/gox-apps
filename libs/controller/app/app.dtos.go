package app

type HealthDto struct{}

type GetRoutesDto struct {
	Length int `form:"length"`
}

type GetRouteDto struct {
	Path string `form:"path" binding:"required"`
}
