package voucher

type CreateVoucherDto struct {
	UserId        string  `json:"userId"        binding:"required"`
	ClientId      string  `json:"clientId"      binding:"required"`
	WorkspaceId   string  `json:"workspaceId"   binding:"required"`
	VoucherTypeId string  `json:"voucherTypeId" binding:"required"`
	Status        *string `json:"status,omitempty"`
}

type UpdateVoucherDto struct {
	Status *string `json:"status,omitempty"`
}
