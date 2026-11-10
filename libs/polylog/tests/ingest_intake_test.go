package tests

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/polylog/app"
	"github.com/thescaffold/gox-apps/libs/polylog/app/event"
	"github.com/thescaffold/gox-apps/libs/polylog/app/source"
	"github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// PLAN U26: POST /ingest and /ingest/batch carry no signature, so they must not
// accept an event for a source that only takes signed webhooks.
func TestOpenIntakeRefusesSourcesThatRequireASignature(t *testing.T) {
	db := polylogDB(t)
	svc := app.NewAppServiceForTest(db)
	t.Setenv("ROOT_USER_ID", "root-user")
	t.Setenv("ROOT_CLIENT_ID", "root-client")
	t.Setenv("ROOT_WORKSPACE_ID", "root-ws")

	key := "signed-" + uuid.NewString()[:8]
	signed := &source.Source{UserId: "root-user", ClientId: "root-client", WorkspaceId: "root-ws", Category: "webhook", Key: &key, TypeId: uuid.NewString(), Name: key,
		Meta: json.RawMessage(`{"signature":{"header":"X-Sig","secretEnv":"POLYLOG_TEST_SIGNING_KEY"}}`)}
	signed.Id = uuid.NewString()
	if err := db.Create(signed).Error; err != nil {
		t.Fatal(err)
	}
	var before, after int64
	db.Model(&event.Event{}).Count(&before)

	for _, it := range []pipeline.Item{
		{Source: signed.Id, Type: "chain.deposit"},
		{Source: key, Type: "chain.deposit"},
		{EntityId: signed.Id, EntityName: "source", Type: "chain.deposit"},
	} {
		if _, err := svc.Ingest(it, ntxctx.NTXContext{}); !errors.Is(err, app.ErrSignature) {
			t.Fatalf("ingest %+v for a signed source: %v, want ErrSignature", it, err)
		}
	}
	if _, err := svc.IngestBatch([]pipeline.Item{{Type: "page.view"}, {Source: signed.Id, Type: "chain.deposit"}}, ntxctx.NTXContext{}); !errors.Is(err, app.ErrSignature) {
		t.Fatalf("a batch with one item for a signed source: %v", err)
	}
	db.Model(&event.Event{}).Count(&after)
	if after != before {
		t.Fatalf("%d refused events were stored", after-before)
	}

	// Anything else still comes in, as browser SDKs expect.
	if _, err := svc.IngestBatch([]pipeline.Item{{Type: "page.view", EntityId: uuid.NewString(), EntityName: "source"}}, ntxctx.NTXContext{}); err != nil {
		t.Fatalf("an unsigned source's batch: %v", err)
	}

	// A deployment that takes signed webhooks only refuses the open intake outright.
	t.Setenv(app.RequireSignatureVar, "true")
	if _, err := svc.Ingest(pipeline.Item{Type: "page.view"}, ntxctx.NTXContext{}); !errors.Is(err, app.ErrSignature) {
		t.Fatalf("with %s: %v", app.RequireSignatureVar, err)
	}
}
