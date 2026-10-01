package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/notification/app"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationtemplate "github.com/thescaffold/gox-apps/libs/notification/app/template"
	"github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// mailServer stands in for the mail provider (Mailgun's API shape).
type mailServer struct {
	*httptest.Server
	mu     sync.Mutex
	bodies []string
	types  []string
	status int
}

func newMailServer(t *testing.T) *mailServer {
	t.Helper()
	m := &mailServer{status: http.StatusOK}
	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		m.mu.Lock()
		m.bodies = append(m.bodies, string(b))
		m.types = append(m.types, r.Header.Get("Content-Type"))
		st := m.status
		m.mu.Unlock()
		w.WriteHeader(st)
		_, _ = w.Write([]byte(`{"id":"x","message":"Queued"}`))
	}))
	t.Cleanup(m.Close)
	return m
}

func (m *mailServer) setStatus(s int) { m.mu.Lock(); m.status = s; m.mu.Unlock() }
func (m *mailServer) sent() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.bodies...)
}

func deliverFixture(t *testing.T) (*app.AppService, *gorm.DB, *mailServer) {
	t.Helper()
	dsn := os.Getenv("NOTIFICATION_TEST_DB_URL")
	if dsn == "" {
		t.Skip("NOTIFICATION_TEST_DB_URL not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for _, m := range app.Migrations {
		if err := m.Run(q); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
	mail := newMailServer(t)
	t.Setenv("GROUP_NAME", "deliver-test")
	t.Setenv("NOTIFICATION_DEFAULT_PROVIDER", "mailgun")
	t.Setenv("MAILGUN_BASE_URL", mail.URL)
	t.Setenv("MAILGUN_DOMAIN", "mg.example.test")
	t.Setenv("MAILGUN_TOKEN", "key-test")
	t.Setenv("MAILGUN_FROM", "Origine <no-reply@mg.example.test>")
	return app.NewAppService(db, provider.NewService(db, "hmac-test-key")), db, mail
}

func seedTemplate(t *testing.T, db *gorm.DB, key string) {
	t.Helper()
	email, web := "<p>{{name}}: {{summary}}</p><a href=\"{{link}}\">Review</a>", "{{summary}}"
	tpl := notificationtemplate.Template{Group: "deliver-test", Key: key, Email: &email, Web: &web}
	tpl.Id = uuid.NewString()
	if err := db.Create(&tpl).Error; err != nil {
		t.Fatal(err)
	}
}

func payload(key string) map[string]any {
	return map[string]any{
		"reference": "ref-" + uuid.NewString(), "key": key, "userId": "u-" + uuid.NewString()[:8],
		"workspaceId": "ws-1", "subject": "Approval needed", "channels": []string{"web", "email"},
		"data": map[string]any{"owner": "ada@example.test", "name": "Ada", "summary": "Deploy <b>v12</b>", "link": "https://app.example.test/a/1"},
	}
}

func TestDeliver_SendsTheEmailAndLeavesAnInAppMessage(t *testing.T) {
	svc, db, mail := deliverFixture(t)
	key := "k-" + uuid.NewString()[:8]
	seedTemplate(t, db, key)
	p := payload(key)

	if err := svc.Deliver(p); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	sent := mail.sent()
	if len(sent) != 1 {
		t.Fatalf("provider got %d requests, want 1", len(sent))
	}
	m := mail
	m.mu.Lock()
	ct := m.types[0]
	m.mu.Unlock()
	if !strings.HasPrefix(ct, "multipart/form-data; boundary=") {
		t.Fatalf("the provider was sent %q, not a real multipart form (Mailgun rejects JSON)", ct)
	}
	for _, want := range []string{"ada@example.test", "Approval needed", "Ada: Deploy &lt;b&gt;v12&lt;/b&gt;", "https://app.example.test/a/1"} {
		if !strings.Contains(sent[0], want) {
			t.Fatalf("email lacks %q:\n%s", want, sent[0])
		}
	}
	var msgs []notificationmessage.Message
	if err := db.Where("user_id = ?", p["userId"]).Find(&msgs).Error; err != nil || len(msgs) != 1 {
		t.Fatalf("in-app messages = %d, %v; want 1", len(msgs), err)
	}
	var l notificationlog.Log
	if err := db.Where("reference = ?", p["reference"]).First(&l).Error; err != nil || l.Status == nil || *l.Status != "success" {
		t.Fatalf("log = %+v, %v; want status success", l, err)
	}
}

func TestDeliver_ReportsAProviderFailureAndRetryingDoesNotDuplicateTheLog(t *testing.T) {
	svc, db, mail := deliverFixture(t)
	key := "k-" + uuid.NewString()[:8]
	seedTemplate(t, db, key)
	p := payload(key)

	mail.setStatus(http.StatusInternalServerError)
	if err := svc.Deliver(p); err == nil {
		t.Fatal("a failed send was reported as delivered")
	}
	mail.setStatus(http.StatusOK)
	if err := svc.Deliver(p); err != nil {
		t.Fatalf("retry: %v", err)
	}
	var n int64
	db.Model(&notificationlog.Log{}).Where("reference = ?", p["reference"]).Count(&n)
	if n != 1 {
		t.Fatalf("%d log rows for one reference, want 1", n)
	}
	var msgs int64
	db.Model(&notificationmessage.Message{}).Where("user_id = ?", p["userId"]).Count(&msgs)
	if msgs != 1 {
		t.Fatalf("a failed-then-retried notification left %d in-app messages, want 1", msgs)
	}
}

func TestDeliver_ADeliveredReferenceIsNotSentTwice(t *testing.T) {
	svc, db, mail := deliverFixture(t)
	key := "k-" + uuid.NewString()[:8]
	seedTemplate(t, db, key)
	p := payload(key)
	for i := 0; i < 3; i++ {
		if err := svc.Deliver(p); err != nil {
			t.Fatalf("deliver %d: %v", i, err)
		}
	}
	if got := len(mail.sent()); got != 1 {
		t.Fatalf("one reference produced %d emails, want 1", got)
	}
	var msgs int64
	db.Model(&notificationmessage.Message{}).Where("user_id = ?", p["userId"]).Count(&msgs)
	if msgs != 1 {
		t.Fatalf("one reference produced %d in-app messages, want 1", msgs)
	}
}

func TestDeliver_AMissingTemplateIsAnErrorNotSilence(t *testing.T) {
	svc, _, mail := deliverFixture(t)
	if err := svc.Deliver(payload("no-such-template-" + uuid.NewString()[:8])); err == nil {
		t.Fatal("a notification with no template was reported as delivered")
	}
	if len(mail.sent()) != 0 {
		t.Fatal("something was sent without a template")
	}
}

func TestDeliver_RejectsAnIncompletePayload(t *testing.T) {
	svc, _, _ := deliverFixture(t)
	for name, p := range map[string]map[string]any{
		"no reference": {"key": "k", "userId": "u", "channels": []string{"email"}},
		"no key":       {"reference": "r", "userId": "u", "channels": []string{"email"}},
		"no user":      {"reference": "r", "key": "k", "channels": []string{"email"}},
	} {
		if err := svc.Deliver(p); err == nil {
			t.Fatalf("%s: accepted", name)
		}
	}
}

func TestService_IsTheRunningAppServiceOnceTheModuleIsDeclared(t *testing.T) {
	before := app.Service()
	m := &app.AppModule{}
	decl := m.Declarations()
	var declared *app.AppService
	for _, d := range decl {
		if s, ok := d.(*app.AppService); ok {
			declared = s
		}
	}
	if declared == nil || app.Service() != declared {
		t.Fatalf("Service() = %p, want the declared service %p (was %p)", app.Service(), declared, before)
	}
}
