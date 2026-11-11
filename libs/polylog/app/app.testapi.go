package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-apps/libs/polylog/app/channel"
	"github.com/thescaffold/gox-apps/libs/polylog/app/event"
	"github.com/thescaffold/gox-apps/libs/polylog/app/source"
	"github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	"gorm.io/gorm"
)

// NewAppServiceForTest builds the ingest side of AppService over db without
// the container.
func NewAppServiceForTest(db *gorm.DB) *AppService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	src := &source.SourceEntity{Entity: (&sql.Entity[source.Source]{}).With(q)}
	ev := &event.EventEntity{Entity: (&sql.Entity[event.Event]{}).With(q)}
	ch := &channel.ChannelEntity{Entity: (&sql.Entity[channel.Channel]{}).With(q)}
	src.OnRegister()
	ev.OnRegister()
	ch.OnRegister()
	return &AppService{pipelineService: pipeline.NewServiceForTest(db), sourceEntity: src, eventEntity: ev, channelEntity: ch}
}
