package formtype

type CreateFormTypeDto struct {
	Name         string  `json:"name"         binding:"required"`
	Desc         *string `json:"desc,omitempty"`
	Detail       *string `json:"detail,omitempty"`
	ThumbnailUrl *string `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string `json:"bannerUrl,omitempty"`
	Tags         *string `json:"tags,omitempty"`
	Meta         *string `json:"meta,omitempty"`
}

type UpdateFormTypeDto struct {
	Name         *string `json:"name,omitempty"`
	Desc         *string `json:"desc,omitempty"`
	Detail       *string `json:"detail,omitempty"`
	ThumbnailUrl *string `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string `json:"bannerUrl,omitempty"`
	Tags         *string `json:"tags,omitempty"`
	Meta         *string `json:"meta,omitempty"`
}
