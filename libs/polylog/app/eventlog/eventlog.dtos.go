package eventlog
import "encoding/json"
type CreateEventLogDto struct {
	EventId  string          `json:"eventId"  binding:"required"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
type UpdateEventLogDto struct{ Status *string `json:"status,omitempty"` }
