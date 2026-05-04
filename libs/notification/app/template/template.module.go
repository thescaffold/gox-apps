package template

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type TemplateModule struct{}

func (m *TemplateModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *TemplateModule) Exports() []any { return []any{&TemplateService{}, &TemplateEntity{}} }

func (m *TemplateModule) Declarations() []any {
	return []any{&TemplateService{}, &TemplateEntity{}}
}
