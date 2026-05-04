package tokenlog

import "github.com/awesome-goose/goose/modules/sql"

type TokenLog struct {
	sql.BaseEntity

	TokenId string  `gorm:"column:token_id;type:varchar(36);not null" json:"tokenId"`
	Value   string  `gorm:"column:value;type:varchar(255);not null"   json:"value"`
	Status  *string `gorm:"column:status;type:varchar(255)"           json:"status,omitempty"`
}

func (TokenLog) TableName() string { return "FussTokenLogs" }

type TokenLogEntity struct {
	*sql.Entity[TokenLog] `inject:""`
}

func (e *TokenLogEntity) OnRegister() {
	e.Hydrate("FussTokenLogs", []string{"token_id", "value"}, nil, nil, nil, nil, nil, "created_at desc")
}
