package app

type HealthDto struct{}

// GetFormDto carries the :id path param for GET /one/:id.
type GetFormDto struct {
	Id string `param:"id" binding:"required"`
}

// SaveFormDto carries :id + free-form body for POST /one/:id.
type SaveFormDto struct {
	Id   string         `param:"id"      binding:"required"`
	Body map[string]any `json:",inline"`
}
