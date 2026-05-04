package log

type CreateLogDto struct {
	JobId  string `json:"jobId"  binding:"required"`
	Output any    `json:"output,omitempty"`
	Status string `json:"status" binding:"required"`
}

type UpdateLogDto struct {
	Output any     `json:"output,omitempty"`
	Status *string `json:"status,omitempty"`
}
