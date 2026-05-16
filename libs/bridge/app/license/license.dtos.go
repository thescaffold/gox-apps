package license

import "time"

// LicenseTypeRef mirrors TS LicenseTypeDto — minimal {id} reference.
type LicenseTypeRef struct {
	Id string `json:"id" binding:"required"`
}

// PreferenceRef mirrors TS PreferenceDto on the License body — env/flag
// overrides flattened onto the licence at create/update time.
type PreferenceRef struct {
	Key   string  `json:"key"            binding:"required"`
	Value *string `json:"value,omitempty"`
	Type  string  `json:"type"           binding:"required"`
}

// PlanTypeRef mirrors TS PlanTypeDto on the License body — per-plan pricing
// rolled into the licence.
type PlanTypeRef struct {
	Name     string   `json:"name"     binding:"required"`
	Desc     *string  `json:"desc,omitempty"`
	Detail   *string  `json:"detail,omitempty"`
	Type     *string  `json:"type,omitempty"`
	Currency *string  `json:"currency,omitempty"`
	Daily    *float64 `json:"daily,omitempty"`
	Weekly   *float64 `json:"weekly,omitempty"`
	Monthly  float64  `json:"monthly"  binding:"required"`
	Yearly   *float64 `json:"yearly,omitempty"`
	Flags    []string `json:"flags,omitempty"`
}

// WebhookRef mirrors TS WebhookDto on the License body — webhook URL +
// subscribed event list.
type WebhookRef struct {
	Url    string   `json:"url"             binding:"required"`
	Events []string `json:"events,omitempty"`
}

// CreateLicenseDto mirrors ntx-apps/libs/bridge/src/api/license/dto/create-license.dto.ts.
// The TS body carries a nested {type, periodType, preferences, planTypes,
// webhooks, amount} shape; the controller morph flattens it onto the License
// entity row, serialising the nested data into the entity's Meta column.
type CreateLicenseDto struct {
	// User/Client/Workspace IDs are injected by the controller morph from
	// request context; they aren't on the TS body but kept here so callers
	// that bypass the morph still have somewhere to set them.
	UserId      string `json:"userId,omitempty"`
	ClientId    string `json:"clientId,omitempty"`
	WorkspaceId string `json:"workspaceId,omitempty"`

	Type        LicenseTypeRef  `json:"type"        binding:"required"`
	TypeId      string          `json:"-"` // populated by morph from Type.Id
	PeriodType  *string         `json:"periodType,omitempty"`
	Preferences []PreferenceRef `json:"preferences,omitempty"`
	PlanTypes   []PlanTypeRef   `json:"planTypes,omitempty"`
	Webhooks    []WebhookRef    `json:"webhooks,omitempty"`
	Amount      *float64        `json:"amount,omitempty"`

	// Meta carries the serialised TS-side nested body (set by the morph) so
	// the AfterCreate/AfterUpdate hooks can reconstruct it for child sync.
	Meta   *string `json:"-"`
	Status *string `json:"-"`

	Token     *string    `json:"-"`
	StartAt   *time.Time `json:"-"`
	RenewedAt *time.Time `json:"-"`
	ExpiredAt *time.Time `json:"-"`
}

// UpdateLicenseDto mirrors TS UpdateLicenseDto. The morph re-serialises any
// supplied nested data into Meta; existing fields stay where they are.
type UpdateLicenseDto struct {
	Type        *LicenseTypeRef `json:"type,omitempty"`
	PeriodType  *string         `json:"periodType,omitempty"`
	Preferences []PreferenceRef `json:"preferences,omitempty"`
	PlanTypes   []PlanTypeRef   `json:"planTypes,omitempty"`
	Webhooks    []WebhookRef    `json:"webhooks,omitempty"`
	Amount      *float64        `json:"amount,omitempty"`

	Meta   *string `json:"-"`
	Status *string `json:"-"`

	Token     *string    `json:"-"`
	StartAt   *time.Time `json:"-"`
	RenewedAt *time.Time `json:"-"`
	ExpiredAt *time.Time `json:"-"`
}
