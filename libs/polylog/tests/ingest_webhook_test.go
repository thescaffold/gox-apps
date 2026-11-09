package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/awesome-goose/goose/io/input"
	"github.com/awesome-goose/goose/modules/sql"
	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/polylog/app"
	"github.com/thescaffold/gox-apps/libs/polylog/app/event"
	"github.com/thescaffold/gox-apps/libs/polylog/app/source"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The ingest routes hand the handler the body exactly as sent, every query
// parameter and every header (PLAN U6: they used to bind none of them).
func TestIngestDtoBindsRawBodyQueriesAndHeaders(t *testing.T) {
	ctx := goosetest.NewMockContext()
	body := []byte(`{"amount": 123456789012345678901234567890}`)
	ctx.MockRequest().WithParams(map[string]string{"id": "src-1"}).
		WithQueries(map[string]string{"type": "chain.deposit", "category": "webhook"}).
		WithHeader("X-Alchemy-Signature", "abc").WithBody(body)
	var dto app.IngestSourceParamDto
	if err := input.NewInput(ctx).Populate(&dto); err != nil {
		t.Fatal(err)
	}
	if dto.ID != "src-1" || string(dto.Payload) != string(body) || dto.Properties["type"] != "chain.deposit" || dto.Headers["X-Alchemy-Signature"] != "abc" {
		t.Fatalf("%+v", dto)
	}
}

func polylogDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("POLYLOG_TEST_DB_URL")
	if dsn == "" {
		t.Skip("POLYLOG_TEST_DB_URL not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for _, m := range app.Migrations {
		if err := m.Run(q); err != nil {
			t.Fatalf("migrate %T: %v", m, err)
		}
	}
	return db
}

func TestSignedSourceIngest(t *testing.T) {
	db := polylogDB(t)
	svc := app.NewAppServiceForTest(db)
	t.Setenv("POLYLOG_TEST_SIGNING_KEY", "whsec-test")
	t.Setenv("ROOT_USER_ID", "root-user")
	t.Setenv("ROOT_CLIENT_ID", "root-client")
	t.Setenv("ROOT_WORKSPACE_ID", "root-ws")

	key := "alchemy-" + uuid.NewString()[:8]
	src := &source.Source{UserId: "root-user", ClientId: "root-client", WorkspaceId: "root-ws", Category: "webhook", Key: &key, TypeId: uuid.NewString(), Name: key,
		Meta: json.RawMessage(`{"signature":{"header":"X-Alchemy-Signature","secretEnv":"POLYLOG_TEST_SIGNING_KEY"}}`)}
	src.Id = uuid.NewString()
	if err := db.Create(src).Error; err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var heard []map[string]any
	events.Default().Subscribe(app.EventIngested, func(_ string, p any) {
		mu.Lock()
		heard = append(heard, p.(map[string]any))
		mu.Unlock()
	})

	body := []byte(`{"webhookId":"wh_1","event":{"activity":[{"rawContract":{"rawValue":"0xde0b6b3a7640000"}}]}}`)
	m := hmac.New(sha256.New, []byte("whsec-test"))
	m.Write(body)
	good := hex.EncodeToString(m.Sum(nil))
	props := map[string]string{"type": "chain.deposit", "category": "webhook"}

	// A bad signature is refused and nothing is stored.
	var before int64
	db.Model(&event.Event{}).Count(&before)
	if _, err := svc.IngestSourceFrom(body, map[string]string{"X-Alchemy-Signature": "00"}, src.Id, props, ntxctx.NTXContext{}); !errors.Is(err, app.ErrSignature) {
		t.Fatalf("bad signature: %v", err)
	}
	var after int64
	db.Model(&event.Event{}).Count(&after)
	if after != before {
		t.Fatal("a refused webhook was stored")
	}

	// A good one (by key or id) is stored with its body and type, and announced.
	out, err := svc.IngestSourceFrom(body, map[string]string{"X-Alchemy-Signature": good, "Authorization": "secret"}, key, props, ntxctx.NTXContext{})
	if err != nil {
		t.Fatal(err)
	}
	ev := out.(*event.Event)
	var stored event.Event
	if err := db.First(&stored, "id = ?", ev.Id).Error; err != nil {
		t.Fatal(err)
	}
	var got, want any
	_ = json.Unmarshal(stored.Payload, &got)
	_ = json.Unmarshal(body, &want)
	gb, _ := json.Marshal(got)
	wb, _ := json.Marshal(want)
	if string(gb) != string(wb) || stored.Type == nil || *stored.Type != "chain.deposit" || stored.Category != "webhook" {
		t.Fatalf("stored %s type %v category %q", stored.Payload, stored.Type, stored.Category)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(heard) != 1 || heard[0]["verified"] != true || heard[0]["body"] != string(body) || heard[0]["type"] != "chain.deposit" {
		t.Fatalf("announcements %v", heard)
	}
	if h := heard[0]["headers"].(map[string]string); h["Authorization"] != "" || h["X-Alchemy-Signature"] != good {
		t.Fatalf("forwarded headers %v", h)
	}
}

func TestUnsignedSourcesStillIngestUnlessRequired(t *testing.T) {
	db := polylogDB(t)
	svc := app.NewAppServiceForTest(db)
	body := []byte("not json")
	out, err := svc.IngestSourceFrom(body, nil, "no-such-source", map[string]string{"type": "x.y"}, ntxctx.NTXContext{})
	if err != nil {
		t.Fatal(err)
	}
	var stored event.Event
	db.First(&stored, "id = ?", out.(*event.Event).Id)
	if string(stored.Payload) != `{"raw": "not json"}` && string(stored.Payload) != `{"raw":"not json"}` {
		t.Fatalf("payload %s", stored.Payload)
	}
	t.Setenv(app.RequireSignatureVar, "true")
	if _, err := svc.IngestSourceFrom(body, nil, "no-such-source", nil, ntxctx.NTXContext{}); !errors.Is(err, app.ErrSignature) {
		t.Fatalf("required signature: %v", err)
	}
}
