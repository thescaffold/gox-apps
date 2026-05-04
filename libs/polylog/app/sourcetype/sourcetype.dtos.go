package sourcetype

type CreateSourceTypeDto struct {
	Category     string  `json:"category"              binding:"required"`
	Key          *string `json:"key,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	Name         string  `json:"name"                  binding:"required"`
	Desc         *string `json:"desc,omitempty"`
	Detail       *string `json:"detail,omitempty"`
	ThumbnailUrl *string `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string `json:"bannerUrl,omitempty"`
	Tags         *string `json:"tags,omitempty"`
	Meta         *string `json:"meta,omitempty"`
}

type UpdateSourceTypeDto struct {
	Category     *string `json:"category,omitempty"`
	Key          *string `json:"key,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	Name         *string `json:"name,omitempty"`
	Desc         *string `json:"desc,omitempty"`
	Detail       *string `json:"detail,omitempty"`
	ThumbnailUrl *string `json:"thumbnailUrl,omitempty"`
	BannerUrl    *string `json:"bannerUrl,omitempty"`
	Tags         *string `json:"tags,omitempty"`
	Meta         *string `json:"meta,omitempty"`
}
