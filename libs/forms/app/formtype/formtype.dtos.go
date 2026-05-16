package formtype

import "encoding/json"

// CreateFormTypeDto mirrors ntx-apps/libs/forms/src/api/form-type/dto/create-form-type.dto.ts.
type CreateFormTypeDto struct {
	Name         string          `json:"name"               binding:"required"`
	Desc         *string         `json:"desc,omitempty"`
	Detail       *string         `json:"detail,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string         `json:"bannerUrl,omitempty"`
	Tags         json.RawMessage `json:"tags,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
}

// UpdateFormTypeDto mirrors ntx-apps/libs/forms/src/api/form-type/dto/update-form-type.dto.ts.
type UpdateFormTypeDto struct {
	Name         *string         `json:"name,omitempty"`
	Desc         *string         `json:"desc,omitempty"`
	Detail       *string         `json:"detail,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string         `json:"bannerUrl,omitempty"`
	Tags         json.RawMessage `json:"tags,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
}
