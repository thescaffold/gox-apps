package app

// PushDto mirrors TS PushDto exactly:
//
//	{queue, job, data, priority, retryLimit, retryDelay, startAt, expireAt,
//	 singleton, frequency}
//
// startAt / expireAt are ISO-8601 strings (Go parses on consumption).
type PushDto struct {
	Queue      string `json:"queue" binding:"required"`
	Job        string `json:"job"   binding:"required"`
	Data       any    `json:"data,omitempty"`
	Priority   int    `json:"priority"`
	RetryLimit int    `json:"retryLimit"`
	RetryDelay int    `json:"retryDelay"`
	// StartAt — ISO-8601; job will not be popped before this instant.
	StartAt string `json:"startAt,omitempty"`
	// ExpireAt — ISO-8601; job will not be popped after this instant.
	ExpireAt string `json:"expireAt,omitempty"`
	// Singleton — when true, only one in-flight instance of this (queue,job) is allowed.
	Singleton bool `json:"singleton,omitempty"`
	// Frequency — when set, the job recurs (e.g. "1h", "daily"). TS reads this verbatim.
	Frequency string `json:"frequency,omitempty"`
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
