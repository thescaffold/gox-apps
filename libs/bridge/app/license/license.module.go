package license

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type LicenseModule struct{}

func (m *LicenseModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *LicenseModule) Exports() []any      { return []any{&LicenseService{}} }
func (m *LicenseModule) Declarations() []any { return []any{&LicenseService{}, &LicenseEntity{}} }
