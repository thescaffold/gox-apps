package pipeline

import (
	"github.com/awesome-goose/goose/types"
	polylogchannel "github.com/thescaffold/gox-apps-polylog/app/channel"
	polyloevent "github.com/thescaffold/gox-apps-polylog/app/event"
	polylogeventlog "github.com/thescaffold/gox-apps-polylog/app/eventlog"
	polylogsink "github.com/thescaffold/gox-apps-polylog/app/sink"
	polylogsinktype "github.com/thescaffold/gox-apps-polylog/app/sinktype"
)

type PipelineModule struct{}

func (m *PipelineModule) Imports() []types.Module {
	return []types.Module{
		&polylogchannel.ChannelModule{},
		&polyloevent.EventModule{},
		&polylogeventlog.EventLogModule{},
		&polylogsink.SinkModule{},
		&polylogsinktype.SinkTypeModule{},
	}
}

func (m *PipelineModule) Exports() []any { return []any{&PipelineService{}} }

func (m *PipelineModule) Declarations() []any {
	return []any{
		&PipelineService{},
		&WebhookSink{},
	}
}
