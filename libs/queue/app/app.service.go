package app

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
)

type AppService struct {
	queueSvc *goqueues.Queue `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

func (s *AppService) Push(queueName, jobName string, data any, config *goqueues.JobConfig) (*goqueues.QueueJob, error) {
	return s.queueSvc.Push(queueName, jobName, data, config)
}

func (s *AppService) Pop(queueName, jobName string) (*goqueues.QueueJob, error) {
	return s.queueSvc.Pop(queueName, jobName)
}

func (s *AppService) Log(jobId, status string, output any) (*goqueues.QueueJob, error) {
	return s.queueSvc.Log(jobId, status, output)
}
