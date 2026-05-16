package voucher

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

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

// VoucherListDto carries the GET /voucher/list query params.
type VoucherListDto struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Scope string            `query:"scope"`
}

// VoucherVerifyDto carries the POST /voucher/verify/:token body + path.
type VoucherVerifyDto struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Token string            `param:"token"`
	Scope string            `json:"scope,omitempty"`
}
