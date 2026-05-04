package job

type CreateJobDto struct {
	QueueId    string `json:"queueId"  binding:"required"`
	Name       string `json:"name"     binding:"required"`
	Priority   int    `json:"priority"`
	RetryLimit int    `json:"retryLimit"`
	RetryDelay int    `json:"retryDelay"`
}

type UpdateJobDto struct {
	Name       *string `json:"name,omitempty"`
	Priority   *int    `json:"priority,omitempty"`
	RetryLimit *int    `json:"retryLimit,omitempty"`
	RetryDelay *int    `json:"retryDelay,omitempty"`
	Status     *string `json:"status,omitempty"`
}
