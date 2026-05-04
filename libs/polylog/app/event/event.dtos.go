package event
import "encoding/json"
type CreateEventDto struct {
	UserId      string          `json:"userId"      binding:"required"`
	ClientId    string          `json:"clientId"    binding:"required"`
	WorkspaceId string          `json:"workspaceId" binding:"required"`
	EntityId    string          `json:"entityId"    binding:"required"`
	EntityName  string          `json:"entityName"  binding:"required"`
	Category    string          `json:"category"    binding:"required"`
	Reference   string          `json:"reference"   binding:"required"`
	Version     string          `json:"version"     binding:"required"`
	Payload     json.RawMessage `json:"payload"     binding:"required"`
	Type        *string         `json:"type,omitempty"`
	Status      *string         `json:"status,omitempty"`
}
type UpdateEventDto struct {
	Status *string `json:"status,omitempty"`
}
