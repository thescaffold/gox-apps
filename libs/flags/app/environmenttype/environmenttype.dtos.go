package environmenttype

import "encoding/json"

// CreateEnvironmentTypeDto mirrors ntx-apps/libs/flags/src/api/environment-type/dto/create-environment-type.dto.ts.
type CreateEnvironmentTypeDto struct {
	Category     string          `json:"category"     binding:"required"`
	Name         string          `json:"name"         binding:"required"`
	Desc         *string         `json:"desc,omitempty"`
	Detail       *string         `json:"detail,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string         `json:"bannerUrl,omitempty"`
	Tags         json.RawMessage `json:"tags,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
}

// UpdateEnvironmentTypeDto mirrors ntx-apps/libs/flags/src/api/environment-type/dto/update-environment-type.dto.ts.
type UpdateEnvironmentTypeDto struct {
	Category     *string         `json:"category,omitempty"`
	Name         *string         `json:"name,omitempty"`
	Desc         *string         `json:"desc,omitempty"`
	Detail       *string         `json:"detail,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string         `json:"bannerUrl,omitempty"`
	Tags         json.RawMessage `json:"tags,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
}
