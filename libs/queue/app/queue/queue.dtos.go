package queue

type CreateQueueDto struct {
	Name     string `json:"name"     binding:"required"`
	Priority int    `json:"priority"`
	Status   string `json:"status,omitempty"`
}

type UpdateQueueDto struct {
	Name     *string `json:"name,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Status   *string `json:"status,omitempty"`
}
