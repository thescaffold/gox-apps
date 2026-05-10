package preference

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type PreferenceModule struct{}

func (m *PreferenceModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *PreferenceModule) Exports() []any { return []any{&PreferenceService{}} }
func (m *PreferenceModule) Declarations() []any {
	return []any{&PreferenceService{}, &PreferenceEntity{}}
}
