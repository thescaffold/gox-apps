package app

import (
	"testing"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	notificationprovider "github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
	queueapp "github.com/thescaffold/gox-apps/libs/queue/app"
)

// Wiring the queue pusher also starts the workers that drain it, once: a
// message pushed onto queue/apps/notification must be sent, not left queued.
func TestOnRegisterStartsTheMessageWorkers(t *testing.T) {
	var got [][]*goqueues.JobHandler
	registerJobs = func(_ *queueapp.AppService, jobs []*goqueues.JobHandler) { got = append(got, jobs) }
	c := &AppController{provider: &notificationprovider.ProviderService{}, queue: &queueapp.AppService{}}
	c.OnRegister()
	c.OnRegister()
	if len(got) != 1 || len(got[0]) != len(Jobs) || got[0][0].Queue != "queue/apps/notification" || got[0][0].Job != "message" {
		t.Fatalf("registered %v", got)
	}
}
