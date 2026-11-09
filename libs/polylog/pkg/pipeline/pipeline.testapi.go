package pipeline

import (
	"github.com/awesome-goose/goose/modules/sql"
	polylogchannel "github.com/thescaffold/gox-apps/libs/polylog/app/channel"
	polyloevent "github.com/thescaffold/gox-apps/libs/polylog/app/event"
	polylogeventlog "github.com/thescaffold/gox-apps/libs/polylog/app/eventlog"
	polylogsink "github.com/thescaffold/gox-apps/libs/polylog/app/sink"
	polylogsinktype "github.com/thescaffold/gox-apps/libs/polylog/app/sinktype"
	"gorm.io/gorm"
)

// NewServiceForTest builds a PipelineService over db without the container
// (no webhook sink: channels with sinks are skipped).
func NewServiceForTest(db *gorm.DB) *PipelineService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	ev := &polyloevent.EventEntity{Entity: (&sql.Entity[polyloevent.Event]{}).With(q)}
	ch := &polylogchannel.ChannelEntity{Entity: (&sql.Entity[polylogchannel.Channel]{}).With(q)}
	el := &polylogeventlog.EventLogEntity{Entity: (&sql.Entity[polylogeventlog.EventLog]{}).With(q)}
	sk := &polylogsink.SinkEntity{Entity: (&sql.Entity[polylogsink.Sink]{}).With(q)}
	st := &polylogsinktype.SinkTypeEntity{Entity: (&sql.Entity[polylogsinktype.SinkType]{}).With(q)}
	ev.OnRegister()
	ch.OnRegister()
	el.OnRegister()
	sk.OnRegister()
	st.OnRegister()
	return &PipelineService{eventEntity: ev, channelEntity: ch, eventlogEntity: el, sinkEntity: sk, sinkTypeEntity: st}
}
