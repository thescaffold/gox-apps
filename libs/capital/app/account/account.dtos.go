package account

import "time"

type CreateAccountDto struct {
	UserId           string     `json:"userId"           binding:"required"`
	ClientId         string     `json:"clientId"         binding:"required"`
	WorkspaceId      string     `json:"workspaceId"      binding:"required"`
	BookBalance      *int       `json:"bookBalance,omitempty"`
	AvailableBalance *int       `json:"availableBalance,omitempty"`
	Currency         string     `json:"currency"         binding:"required"`
	Reference        string     `json:"reference"        binding:"required"`
	Desc             *string    `json:"desc,omitempty"`
	Label            *string    `json:"label,omitempty"`
	Type             *string    `json:"type,omitempty"`
	Number           *string    `json:"number,omitempty"`
	DailyLimit       *int       `json:"dailyLimit,omitempty"`
	MonthlyLimit     *int       `json:"monthlyLimit,omitempty"`
	Status           *string    `json:"status,omitempty"`
	SuspendedAt      *time.Time `json:"suspendedAt,omitempty"`
	ClosedAt         *time.Time `json:"closedAt,omitempty"`
}

type UpdateAccountDto struct {
	BookBalance      *int       `json:"bookBalance,omitempty"`
	AvailableBalance *int       `json:"availableBalance,omitempty"`
	Desc             *string    `json:"desc,omitempty"`
	Label            *string    `json:"label,omitempty"`
	Type             *string    `json:"type,omitempty"`
	Number           *string    `json:"number,omitempty"`
	DailyLimit       *int       `json:"dailyLimit,omitempty"`
	MonthlyLimit     *int       `json:"monthlyLimit,omitempty"`
	Status           *string    `json:"status,omitempty"`
	SuspendedAt      *time.Time `json:"suspendedAt,omitempty"`
	ClosedAt         *time.Time `json:"closedAt,omitempty"`
}
