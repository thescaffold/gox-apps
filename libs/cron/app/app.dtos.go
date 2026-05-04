package app

type RegisterDto struct {
	Group      string `json:"group"   binding:"required"`
	Name       string `json:"name"    binding:"required"`
	Pattern    string `json:"pattern" binding:"required"`
	Priority   int    `json:"priority"`
	RetryLimit int    `json:"retryLimit"`
	RetryDelay int    `json:"retryDelay"`
}

type SelectDto struct {
	Group string `form:"group" binding:"required"`
	Name  string `form:"name"  binding:"required"`
}

type LogDto struct {
	JobId  string `json:"jobId"  binding:"required"`
	Output any    `json:"output,omitempty"`
	Status string `json:"status" binding:"required"`
}

type HealthDto struct{}
