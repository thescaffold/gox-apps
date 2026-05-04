package history

import "github.com/awesome-goose/goose/modules/sql"

type History struct {
	sql.BaseEntity

	UserId      string  `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string  `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string  `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	Type        string  `gorm:"column:type;type:varchar(255);default:'global'" json:"type"`
	Query       string  `gorm:"column:query;type:varchar(255);not null"       json:"query"`
	Status      *string `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (History) TableName() string { return "FussHistories" }

type HistoryEntity struct {
	*sql.Entity[History] `inject:""`
}

func (e *HistoryEntity) OnRegister() {
	e.Hydrate("FussHistories", []string{"user_id", "query", "type"}, nil, nil, nil, nil, nil, "created_at desc")
}
