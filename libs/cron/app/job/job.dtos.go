package job

type CreateJobDto struct {
	Group      string `json:"group"   binding:"required"`
	Name       string `json:"name"    binding:"required"`
	Pattern    string `json:"pattern" binding:"required"`
	Priority   int    `json:"priority"`
	RetryLimit int    `json:"retryLimit"`
	RetryDelay int    `json:"retryDelay"`
}

type UpdateJobDto struct {
	Pattern    *string `json:"pattern,omitempty"`
	Priority   *int    `json:"priority,omitempty"`
	RetryLimit *int    `json:"retryLimit,omitempty"`
	RetryDelay *int    `json:"retryDelay,omitempty"`
	Status     *string `json:"status,omitempty"`
}
