package summary

type CreateSummaryDto struct {
	ServiceId string  `json:"serviceId" binding:"required"`
	Type      string  `json:"type"      binding:"required"`
	Received  int     `json:"received"`
	Measure   float64 `json:"measure"`
	Note      *string `json:"note,omitempty"`
}

type UpdateSummaryDto struct {
	ServiceId *string  `json:"serviceId,omitempty"`
	Type      *string  `json:"type,omitempty"`
	Received  *int     `json:"received,omitempty"`
	Measure   *float64 `json:"measure,omitempty"`
	Note      *string  `json:"note,omitempty"`
}
