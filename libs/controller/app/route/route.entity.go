package route

import "github.com/awesome-goose/goose/modules/sql"

type Route struct {
	sql.BaseEntity

	Group    string  `gorm:"column:group;type:varchar(255);not null"    json:"group"`
	Service  string  `gorm:"column:service;type:varchar(255);not null"  json:"service"`
	Type     string  `gorm:"column:type;type:varchar(255);default:'http'" json:"type"`
	Name     string  `gorm:"column:name;type:varchar(255);not null"     json:"name"`
	Desc     *string `gorm:"column:desc;type:varchar(255)"              json:"desc,omitempty"`
	Upstream string  `gorm:"column:upstream;type:varchar(255);not null" json:"upstream"`
	Status   *string `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (Route) TableName() string { return "ControllerRoutes" }

type RouteEntity struct {
	*sql.Entity[Route] `inject:""`
}

func (e *RouteEntity) OnRegister() {
	// Mirrors TS route.controller.ts:19 searchable = ['name','desc'].
	e.Hydrate("ControllerRoutes", []string{"name", "desc"}, nil, nil, nil, nil, nil, "created_at desc")
}
