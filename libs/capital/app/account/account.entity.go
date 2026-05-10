package account

import (
	"github.com/awesome-goose/goose/modules/sql"
	"time"
)

type Account struct {
	sql.BaseEntity
	UserId           string     `gorm:"column:user_id;type:varchar(36);not null"       json:"userId"`
	ClientId         string     `gorm:"column:client_id;type:varchar(255);not null"    json:"clientId"`
	WorkspaceId      string     `gorm:"column:workspace_id;type:varchar(36);not null"  json:"workspaceId"`
	BookBalance      int        `gorm:"column:book_balance"                            json:"bookBalance"`
	AvailableBalance int        `gorm:"column:available_balance"                       json:"availableBalance"`
	Currency         string     `gorm:"column:currency;type:varchar(255);not null"     json:"currency"`
	Reference        string     `gorm:"column:reference;type:varchar(255);not null"    json:"reference"`
	Desc             *string    `gorm:"column:desc;type:varchar(255)"                  json:"desc,omitempty"`
	Label            *string    `gorm:"column:label;type:varchar(255)"                 json:"label,omitempty"`
	Type             *string    `gorm:"column:type;type:varchar(255)"                  json:"type,omitempty"`
	Number           *string    `gorm:"column:number;type:varchar(255)"                json:"number,omitempty"`
	DailyLimit       *int       `gorm:"column:daily_limit"                             json:"dailyLimit,omitempty"`
	MonthlyLimit     *int       `gorm:"column:monthly_limit"                           json:"monthlyLimit,omitempty"`
	Status           *string    `gorm:"column:status;type:varchar(255)"                json:"status,omitempty"`
	SuspendedAt      *time.Time `gorm:"column:suspended_at;type:timestamp"             json:"suspendedAt,omitempty"`
	ClosedAt         *time.Time `gorm:"column:closed_at;type:timestamp"                json:"closedAt,omitempty"`
}

func (Account) TableName() string { return "CapitalAccounts" }

type AccountEntity struct {
	*sql.Entity[Account] `inject:""`
}

func (e *AccountEntity) OnRegister() {
	e.Hydrate("CapitalAccounts", []string{"user_id", "workspace_id", "currency", "reference"}, nil, nil, nil, nil, nil, "created_at desc")
}
