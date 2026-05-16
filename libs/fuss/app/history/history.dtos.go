package history

// CreateHistoryDto mirrors ntx-apps/libs/fuss/src/api/history/dto/create-history.dto.ts.
// userId / clientId / workspaceId are NOT submitted by clients — TS adds them
// in HistoryController.morphs.beforeCreate via a payload spread. They live on
// the gox DTO without `binding:"required"` so the morph can populate them.
type CreateHistoryDto struct {
	UserId      string  `json:"userId,omitempty"`
	ClientId    string  `json:"clientId,omitempty"`
	WorkspaceId string  `json:"workspaceId,omitempty"`
	Type        string  `json:"type,omitempty"`
	Query       string  `json:"query"             binding:"required"`
	Status      *string `json:"status,omitempty"`
}

// UpdateHistoryDto mirrors ntx-apps/libs/fuss/src/api/history/dto/update-history.dto.ts.
type UpdateHistoryDto struct {
	Type   *string `json:"type,omitempty"`
	Query  *string `json:"query,omitempty"`
	Status *string `json:"status,omitempty"`
}
