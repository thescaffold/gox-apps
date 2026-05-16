package flag

import "encoding/json"

// CreateFlagDto mirrors ntx-apps/libs/flags/src/api/flag/dto/create-flag.dto.ts.
// userId / clientId / workspaceId are NOT submitted by clients — TS adds them
// via morphs.beforeCreate spread; the gox equivalent populates them in the
// same beforeCreate morph so the new copyAny flows them onto the Flag row.
type CreateFlagDto struct {
	UserId        string          `json:"userId,omitempty"`
	ClientId      string          `json:"clientId,omitempty"`
	WorkspaceId   string          `json:"workspaceId,omitempty"`
	EnvironmentId string          `json:"environmentId" binding:"required"`
	Name          string          `json:"name"          binding:"required"`
	Limit         int             `json:"limit"         binding:"required"`
	Priority      int             `json:"priority"      binding:"required"`
	Level         string          `json:"level"         binding:"required"`
	Meta          json.RawMessage `json:"meta"          binding:"required"`
	Status        *string         `json:"status,omitempty"`
}

// UpdateFlagDto mirrors ntx-apps/libs/flags/src/api/flag/dto/update-flag.dto.ts.
type UpdateFlagDto struct {
	EnvironmentId *string         `json:"environmentId,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Limit         *int            `json:"limit,omitempty"`
	Priority      *int            `json:"priority,omitempty"`
	Level         *string         `json:"level,omitempty"`
	Meta          json.RawMessage `json:"meta,omitempty"`
	Status        *string         `json:"status,omitempty"`
}
