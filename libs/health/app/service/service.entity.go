package service

import "github.com/awesome-goose/goose/modules/sql"

type Service struct {
	sql.BaseEntity

	Name   string  `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Desc   *string `gorm:"column:desc;type:varchar(255)"          json:"desc,omitempty"`
	Type   *string `gorm:"column:type;type:varchar(255)"          json:"type,omitempty"`
	State  *string `gorm:"column:state;type:varchar(255)"         json:"state,omitempty"`
	Status *string `gorm:"column:status;type:varchar(255)"        json:"status,omitempty"`
}

func (Service) TableName() string { return "HealthServices" }

type ServiceEntity struct {
	*sql.Entity[Service] `inject:""`
}

func (e *ServiceEntity) OnRegister() {
	// Mirrors TS service.controller.ts:19 searchable = ['name','desc','state'].
	e.Hydrate("HealthServices", []string{"name", "desc", "state"}, nil, nil, nil, nil, nil, "created_at desc")
}
