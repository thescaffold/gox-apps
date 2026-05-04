package paymentlog
import "encoding/json"
type CreatePaymentLogDto struct {
	PaymentId string          `json:"paymentId" binding:"required"`
	Type      string          `json:"type"      binding:"required"`
	Request   json.RawMessage `json:"request"   binding:"required"`
	Response  json.RawMessage `json:"response,omitempty"`
	Status    *string         `json:"status,omitempty"`
}
type UpdatePaymentLogDto struct{ Status *string `json:"status,omitempty"` }
