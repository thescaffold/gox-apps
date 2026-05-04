package app

type PushDto struct {
	Queue      string `json:"queue" binding:"required"`
	Job        string `json:"job"   binding:"required"`
	Data       any    `json:"data,omitempty"`
	Priority   int    `json:"priority"`
	RetryLimit int    `json:"retryLimit"`
	RetryDelay int    `json:"retryDelay"`
}

type PopDto struct {
	Queue string `form:"queue" binding:"required"`
	Job   string `form:"job"   binding:"required"`
}

type LogDto struct {
	JobId  string `json:"jobId"  binding:"required"`
	Output any    `json:"output,omitempty"`
	Status string `json:"status" binding:"required"`
}

type HealthDto struct{}
