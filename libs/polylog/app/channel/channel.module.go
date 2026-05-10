package channel

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type ChannelModule struct{}

func (m *ChannelModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *ChannelModule) Exports() []any      { return []any{&ChannelService{}, &ChannelEntity{}} }
func (m *ChannelModule) Declarations() []any { return []any{&ChannelService{}, &ChannelEntity{}} }
