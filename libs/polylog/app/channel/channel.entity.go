package channel

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Channel struct {
	sql.BaseEntity
	UserId      string          `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string          `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string          `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	Category    string          `gorm:"column:category;type:varchar(255);not null"    json:"category"`
	Key         *string         `gorm:"column:key;type:varchar(255)"                  json:"key,omitempty"`
	Visibility  *string         `gorm:"column:visibility;type:varchar(255)"           json:"visibility,omitempty"`
	Type        *string         `gorm:"column:type;type:varchar(255)"                 json:"type,omitempty"`
	SourceId    string          `gorm:"column:source_id;type:varchar(36);not null"    json:"sourceId"`
	SinkId      *string         `gorm:"column:sink_id;type:varchar(36)"               json:"sinkId,omitempty"`
	Tags        json.RawMessage `gorm:"column:tags;type:jsonb"                        json:"tags,omitempty"`
	Meta        json.RawMessage `gorm:"column:meta;type:jsonb"                        json:"meta,omitempty"`
	Status      *string         `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Channel) TableName() string { return "PolylogChannels" }

type ChannelEntity struct {
	*sql.Entity[Channel] `inject:""`
}

func (e *ChannelEntity) OnRegister() {
	e.Hydrate("PolylogChannels", []string{"user_id", "workspace_id", "category", "source_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
