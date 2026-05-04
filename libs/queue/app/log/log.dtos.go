package log

type CreateLogDto struct {
	QueueId string `json:"queueId" binding:"required"`
	JobId   string `json:"jobId"   binding:"required"`
	Output  any    `json:"output,omitempty"`
	Status  string `json:"status"  binding:"required"`
}

type UpdateLogDto struct {
	Output any     `json:"output,omitempty"`
	Status *string `json:"status,omitempty"`
}
