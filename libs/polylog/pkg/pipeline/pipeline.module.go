package pipeline

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/types"
	polylogchannel "github.com/thescaffold/gox-apps/libs/polylog/app/channel"
	polyloevent "github.com/thescaffold/gox-apps/libs/polylog/app/event"
	polylogeventlog "github.com/thescaffold/gox-apps/libs/polylog/app/eventlog"
	polylogsink "github.com/thescaffold/gox-apps/libs/polylog/app/sink"
	polylogsinktype "github.com/thescaffold/gox-apps/libs/polylog/app/sinktype"
)

type PipelineModule struct{}

func (m *PipelineModule) Imports() []types.Module {
	return []types.Module{
		// The pipeline needs these modules' services and entities; their routes
		// are registered once by polylog's own AppModule, not again from here.
		router.Silent(&polylogchannel.ChannelModule{}),
		router.Silent(&polyloevent.EventModule{}),
		router.Silent(&polylogeventlog.EventLogModule{}),
		router.Silent(&polylogsink.SinkModule{}),
		router.Silent(&polylogsinktype.SinkTypeModule{}),
	}
}

func (m *PipelineModule) Exports() []any { return []any{&PipelineService{}} }

func (m *PipelineModule) Declarations() []any {
	return []any{
		&PipelineService{},
		&WebhookSink{},
	}
}
