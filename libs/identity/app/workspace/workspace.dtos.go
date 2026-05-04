package workspace

import "encoding/json"

type CreateWorkspaceDto struct {
	UserId    string          `json:"userId"    binding:"required"`
	ClientId  string          `json:"clientId"  binding:"required"`
	Name      string          `json:"name"      binding:"required"`
	Desc      *string         `json:"desc,omitempty"`
	Slug      string          `json:"slug"      binding:"required"`
	LogoUrl   *string         `json:"logoUrl,omitempty"`
	BannerUrl *string         `json:"bannerUrl,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}

type UpdateWorkspaceDto struct {
	Name      *string         `json:"name,omitempty"`
	Desc      *string         `json:"desc,omitempty"`
	Slug      *string         `json:"slug,omitempty"`
	LogoUrl   *string         `json:"logoUrl,omitempty"`
	BannerUrl *string         `json:"bannerUrl,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Status    *string         `json:"status,omitempty"`
}
