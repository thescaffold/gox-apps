package invite

import (
	"time"
	"github.com/awesome-goose/goose/modules/sql"
)

type Invite struct {
	sql.BaseEntity
	UserId      string     `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string     `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId string     `gorm:"column:workspace_id;type:varchar(36);not null" json:"workspaceId"`
	Email       string     `gorm:"column:email;type:varchar(255);not null"       json:"email"`
	RoleId      *string    `gorm:"column:role_id;type:varchar(36)"               json:"roleId,omitempty"`
	Token       string     `gorm:"column:token;type:varchar(255);not null"       json:"token"`
	ExpiresAt   *time.Time `gorm:"column:expires_at;type:timestamp"              json:"expiresAt,omitempty"`
	Status      *string    `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}
func (Invite) TableName() string { return "IdentityInvites" }
type InviteEntity struct{ *sql.Entity[Invite] `inject:""` }
func (e *InviteEntity) OnRegister() {
	e.Hydrate("IdentityInvites", []string{"user_id", "client_id", "workspace_id", "email"}, nil, nil, nil, nil, nil, "created_at desc")
}
