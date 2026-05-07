package batch

import (
	"github.com/awesome-goose/goose/types"
	auditlog "github.com/thescaffold/gox-apps-audit/app/log"
)

type BatchModule struct{}

func (m *BatchModule) Imports() []types.Module {
	return []types.Module{
		&auditlog.LogModule{},
	}
}

func (m *BatchModule) Exports() []any { return nil }

func (m *BatchModule) Declarations() []any {
	return []any{&SummaryBatch{}}
}
