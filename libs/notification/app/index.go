package app

import (
	"encoding/json"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps/libs/notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps/libs/notification/app/template"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "notification"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/notification/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {"*:*:apps:notification:subscription:*:*"},
	"member": {
		"member:create:apps:notification:*:*:workspace",
		"member:update:apps:notification:*:*:self",
		"member:read:apps:notification:*:*:workspace",
		"member:delete:apps:notification:*:*:self",
	},
	"admin":         {"admin:*:apps:notification:*:*:workspace"},
	"global-member": {"global-member:*:apps:notification:*:*:global"},
	"global-admin":  {"global-admin:*:apps:notification:*:*:global"},
}

// notificationAppSvc is set by AppModule.Declarations() so subscription
// handlers can drive typed AppService methods after DI.
var notificationAppSvc *AppService

func payloadOf(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		if p, ok := m["payload"].(map[string]any); ok {
			return p
		}
		return m
	}
	return nil
}

// handleMessageNew mirrors TS subscription 'apps.notification.message.new':
// persist a Log row (with publishAt/expireAt defaults) and push a queue job
// to dispatch via ProviderService. The queue push is best-effort — when the
// queue infra isn't wired the Log row is still persisted for later retry.
func handleMessageNew(_ string, raw any) {
	if notificationAppSvc == nil {
		return
	}
	p := payloadOf(raw)
	if p == nil {
		return
	}
	ref, _ := p["reference"].(string)
	key, _ := p["key"].(string)
	if ref == "" || key == "" {
		return
	}
	notificationAppSvc.OnMessageNew(p)
}

// handleAttributeTypeUpdate mirrors TS subscription
// 'apps.identity.attribute.type.update': upserts a Rule with
// (group='subscription', key=user.ref).
func handleAttributeTypeUpdate(_ string, raw any) {
	if notificationAppSvc == nil {
		return
	}
	p := payloadOf(raw)
	if p == nil {
		return
	}
	typeName, _ := p["type"].(string)
	if typeName != "notification" {
		return
	}
	user, _ := p["user"].(map[string]any)
	ref, _ := user["ref"].(string)
	if ref == "" {
		return
	}
	rulesRaw, _ := p["attributes"]
	var rules json.RawMessage
	if rulesRaw != nil {
		if b, err := json.Marshal(rulesRaw); err == nil {
			rules = b
		}
	}
	_, _ = notificationAppSvc.UpdateSubscription(ref, rules)
}

// Subscriptions mirrors TS subscriptions list. Handlers dispatch through the
// resolved AppService singleton.
var Subscriptions = map[string]events.EventHandler{
	"apps.notification.message.new":       handleMessageNew,
	"apps.identity.attribute.type.update": handleAttributeTypeUpdate,
}

// Top-level exports mirroring ntx-apps/libs/notification/src/index.ts.
var (
	Entities = []any{
		notificationlog.Log{},
		notificationrule.Rule{},
		notificationtemplate.Template{},
		notificationmessage.Message{},
	}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/notification.yaml"}
	Crons           = []any{}
)

// Jobs mirrors ntx-apps/libs/notification/src/index.ts jobs array. The
// `queue/apps/notification/message` consumer hydrates the persisted Log row
// from job.data.reference and dispatches it through the ProviderService.
// Host code: queueApp.RegisterJob(notification.Jobs[0]).
var Jobs = []*goqueues.JobHandler{
	goqueues.NewSimpleHandler("queue/apps/notification", "message", func(job *goqueues.QueueJob) (any, error) {
		if notificationAppSvc == nil {
			return nil, nil
		}
		return notificationAppSvc.OnMessageJob(job)
	}),
}
