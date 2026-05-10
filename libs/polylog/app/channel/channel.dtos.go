package channel

type CreateChannelDto struct {
	UserId      string  `json:"userId"      binding:"required"`
	ClientId    string  `json:"clientId"    binding:"required"`
	WorkspaceId string  `json:"workspaceId" binding:"required"`
	Category    string  `json:"category"    binding:"required"`
	SourceId    string  `json:"sourceId"    binding:"required"`
	SinkId      *string `json:"sinkId,omitempty"`
	Type        *string `json:"type,omitempty"`
	Status      *string `json:"status,omitempty"`
}
type UpdateChannelDto struct {
	SinkId *string `json:"sinkId,omitempty"`
	Status *string `json:"status,omitempty"`
}
