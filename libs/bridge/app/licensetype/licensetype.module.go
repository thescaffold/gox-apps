package licensetype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type LicenseTypeModule struct{}

func (m *LicenseTypeModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *LicenseTypeModule) Exports() []any { return []any{&LicenseTypeService{}} }

func (m *LicenseTypeModule) Declarations() []any {
	return []any{&LicenseTypeService{}, &LicenseTypeEntity{}}
}
