package license

import "time"

type CreateLicenseDto struct {
	UserId      string     `json:"userId"       binding:"required"`
	ClientId    string     `json:"clientId"     binding:"required"`
	WorkspaceId string     `json:"workspaceId"  binding:"required"`
	TypeId      string     `json:"typeId"       binding:"required"`
	PeriodType  *string    `json:"periodType,omitempty"`
	Token       *string    `json:"token,omitempty"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	RenewedAt   *time.Time `json:"renewedAt,omitempty"`
	ExpiredAt   *time.Time `json:"expiredAt,omitempty"`
	Meta        *string    `json:"meta,omitempty"`
	Status      *string    `json:"status,omitempty"`
}

type UpdateLicenseDto struct {
	PeriodType *string    `json:"periodType,omitempty"`
	Token      *string    `json:"token,omitempty"`
	StartAt    *time.Time `json:"startAt,omitempty"`
	RenewedAt  *time.Time `json:"renewedAt,omitempty"`
	ExpiredAt  *time.Time `json:"expiredAt,omitempty"`
	Meta       *string    `json:"meta,omitempty"`
	Status     *string    `json:"status,omitempty"`
}
