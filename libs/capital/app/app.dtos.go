package app

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// PaymentType / ProviderName mirror ntx-apps/libs/capital/src/app.dto.ts enums.
type PaymentType string

const (
	PaymentTypeCharge     PaymentType = "charge"
	PaymentTypeVerify     PaymentType = "verify"
	PaymentTypeWallet     PaymentType = "wallet"
	PaymentTypeRecordOnly PaymentType = "record_only"
)

type ProviderName string

const (
	ProviderPaystack    ProviderName = "paystack"
	ProviderFlutterwave ProviderName = "flutterwave"
	ProviderStripe      ProviderName = "stripe"
)

// HelloDto carries request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// ReferenceDto carries the GET /:type/reference/:reference path params.
type ReferenceDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Type      string            `param:"type"`
	Reference string            `param:"reference"`
}

// ProvidersDto carries GET /providers — no body params besides context.
type ProvidersDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// DefaultProviderDto / InitDto share the same fields — kept as separate types
// to mirror TS naming.
type DefaultProviderDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Amount    float64           `json:"amount"     binding:"required"`
	FirstName string            `json:"firstName"  binding:"required"`
	LastName  string            `json:"lastName"   binding:"required"`
	Meta      json.RawMessage   `json:"meta,omitempty"`
}

// InitDto matches DefaultProviderDto field-for-field; TS InitDto is preferred.
type InitDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Amount    float64           `json:"amount"     binding:"required"`
	FirstName string            `json:"firstName"  binding:"required"`
	LastName  string            `json:"lastName"   binding:"required"`
	Meta      json.RawMessage   `json:"meta,omitempty"`
}

// VerifyPaymentDto carries POST /verify/:type body + path.
type VerifyPaymentDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Type      string            `param:"type"`
	Amount    float64           `json:"amount"    binding:"required"`
	Currency  string            `json:"currency"  binding:"required"`
	Reference string            `json:"reference" binding:"required"`
}

// ChargePaymentDto carries POST /:type/charge/:providerId body + path.
type ChargePaymentDto struct {
	Ctx        ntxctx.NTXContext `context:"ntx"`
	Type       string            `param:"type"`
	ProviderId string            `param:"providerId"`
	Amount     float64           `json:"amount"   binding:"required"`
	Currency   string            `json:"currency" binding:"required"`
	Meta       json.RawMessage   `json:"meta,omitempty"`
}

// PayDto carries POST /pay body.
type PayDto struct {
	Ctx        ntxctx.NTXContext `context:"ntx"`
	ProviderId *string           `json:"providerId,omitempty"`
}

// UpgradeUserPlanDto carries POST /user-plan/compute-upgrade body.
type UpgradeUserPlanDto struct {
	Ctx        ntxctx.NTXContext `context:"ntx"`
	PlanTypeId *string           `json:"planTypeId,omitempty"`
	PeriodType string            `json:"periodType" binding:"required"`
}

// UsageStartDto carries POST /usage/start and /usage/update body.
type UsageStartDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `json:"userId"      binding:"required"`
	ClientId    string            `json:"clientId"    binding:"required"`
	WorkspaceId string            `json:"workspaceId" binding:"required"`
	EntityName  string            `json:"entityName"  binding:"required"`
	EntityId    string            `json:"entityId"    binding:"required"`
	Quantity    *string           `json:"quantity,omitempty"`
	StartAt     *string           `json:"startAt,omitempty"`
	Meta        json.RawMessage   `json:"meta,omitempty"`
}

// UsageStopDto carries POST /usage/stop body.
type UsageStopDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `json:"userId"      binding:"required"`
	ClientId    string            `json:"clientId"    binding:"required"`
	WorkspaceId string            `json:"workspaceId" binding:"required"`
	EntityName  string            `json:"entityName"  binding:"required"`
	EntityId    string            `json:"entityId"    binding:"required"`
	StopAt      *string           `json:"stopAt,omitempty"`
}

// UserPlanDto carries GET /user-plan query params.
type UserPlanDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `query:"userId"`
	ClientId    string            `query:"clientId"`
	WorkspaceId string            `query:"workspaceId"`
}

// CapitalPlanChangeDto carries POST /plan/change body.
type CapitalPlanChangeDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `json:"userId"      binding:"required"`
	WorkspaceId string            `json:"workspaceId" binding:"required"`
	Currency    string            `json:"currency"    binding:"required"`
	PlanType    string            `json:"planType"    binding:"required"`
	PeriodType  string            `json:"periodType"  binding:"required"`
}

// CapitalPlanSubscribeDto carries POST /plan/subscribe body.
type CapitalPlanSubscribeDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `json:"userId"      binding:"required"`
	WorkspaceId string            `json:"workspaceId" binding:"required"`
	Amount      float64           `json:"amount"      binding:"required"`
	Currency    string            `json:"currency"    binding:"required"`
}

// StatusDto / DebtDto / AccrualDto / etc share the bare context-only shape.
type StatusDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}
