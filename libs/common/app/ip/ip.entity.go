package ip

import "github.com/awesome-goose/goose/modules/sql"

type Ip struct {
	sql.BaseEntity

	Value   string  `gorm:"column:value;type:varchar(255);not null;uniqueIndex:idx_common_ips_value" json:"value"`
	Country string  `gorm:"column:country;type:varchar(255);not null" json:"country"`
	City    *string `gorm:"column:city;type:varchar(255)"             json:"city,omitempty"`
	Status  *string `gorm:"column:status;type:varchar(255)"           json:"status,omitempty"`
}

func (Ip) TableName() string { return "CommonIps" }

type IpEntity struct {
	*sql.Entity[Ip] `inject:""`
}

func (e *IpEntity) OnRegister() {
	e.Hydrate("CommonIps", []string{"value", "country", "city"}, nil, nil, nil, nil, nil, "created_at desc")
}
