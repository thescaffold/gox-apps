package devicesession

import "time"

type CreateDeviceSessionDto struct {
	DeviceId  string     `json:"deviceId"  binding:"required"`
	UserId    string     `json:"userId"    binding:"required"`
	Token     string     `json:"token"     binding:"required"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Status    *string    `json:"status,omitempty"`
}

type UpdateDeviceSessionDto struct {
	Token     *string    `json:"token,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Status    *string    `json:"status,omitempty"`
}
