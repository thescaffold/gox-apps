package app

type HealthDto struct{}

type CheckDto struct {
	Id string `param:"id" binding:"required"`
}
