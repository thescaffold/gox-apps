package apikey

import (
	"context"
	"os"
	"sync"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/apikey"
)

// ApiKeyModule serves /keys: issue, list, revoke and rotate a workspace's API keys.
type ApiKeyModule struct{}

func (m *ApiKeyModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}

func (m *ApiKeyModule) Exports() []any { return []any{&ApiKeyService{}} }

func (m *ApiKeyModule) Declarations() []any {
	return []any{&ApiKeyService{}, &ApiKeyController{}}
}

// ApiKeyService owns the key service for the process. The brand comes from
// IDENTITY_KEY_BRAND (for example "lgr"); the host may set LiveAllowed to gate
// live keys behind a go-live review.
type ApiKeyService struct {
	db *sql.Db `inject:""`

	once sync.Once
	svc  *apikey.Service

	// LiveAllowed, when set before first use, gates live keys.
	LiveAllowed func(ctx context.Context, clientID, workspaceID string) bool
}

func (s *ApiKeyService) Keys() *apikey.Service {
	s.once.Do(func() {
		s.svc = &apikey.Service{DB: s.db.DB, Brand: os.Getenv("IDENTITY_KEY_BRAND"), LiveAllowed: s.LiveAllowed}
	})
	return s.svc
}

// Use replaces the underlying service (hosts and tests).
func (s *ApiKeyService) Use(svc *apikey.Service) { s.once.Do(func() {}); s.svc = svc }
