package org

import (
	"sync"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	orgsvc "github.com/thescaffold/gox-apps/libs/identity/pkg/org"
)

// OrgModule serves /org: a workspace's teams, members, custom roles and who holds them.
type OrgModule struct{}

func (m *OrgModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *OrgModule) Exports() []any { return []any{&OrgService{}} }

func (m *OrgModule) Declarations() []any {
	return []any{&OrgService{}, &OrgController{}}
}

// OrgService owns the organisation service for the process.
type OrgService struct {
	db *sql.Db `inject:""`

	once sync.Once
	svc  *orgsvc.Service
}

func (s *OrgService) Org() *orgsvc.Service {
	s.once.Do(func() { s.svc = orgsvc.New(s.db.DB) })
	return s.svc
}

// Use replaces the underlying service (tests).
func (s *OrgService) Use(svc *orgsvc.Service) { s.once.Do(func() {}); s.svc = svc }
