package app

import "github.com/thescaffold/gox-packages/libs/core/events"

const Name = "figs"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {"guest:read:apps:figs:open:*:*"},
	"member":        {"member:create:apps:figs:*:*:workspace", "member:update:apps:figs:*:*:self", "member:read:apps:figs:*:*:workspace", "member:delete:apps:figs:*:*:self"},
	"admin":         {"admin:*:apps:figs:*:*:workspace"},
	"global-member": {"global-member:*:apps:figs:*:*:global"},
	"global-admin":  {"global-admin:*:apps:figs:*:*:global"},
}

// appSvc is set by AppModule.Declarations so Subscriptions handlers can call it after DI.
var figsAppSvc *AppService

func handleFigsEvent(_ string, payload any) {
	if figsAppSvc == nil {
		return
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return
	}
	p := &Payload{
		Meta:   toMap(m["meta"]),
		Input:  toMap(m["input"]),
		Output: toMap(m["output"]),
	}
	f, err := figsAppSvc.Init(p)
	if err != nil || f == nil {
		return
	}
	_, conv, st, err := figsAppSvc.Process(f)
	if err != nil || conv == nil || st == nil {
		return
	}
	figsAppSvc.UpdateFile(f, conv, st) //nolint:errcheck
}

func toMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

// Subscriptions maps event patterns to handlers for this app.
var Subscriptions = map[string]events.EventHandler{
	"apps.figs.message.new": handleFigsEvent,
}

// Top-level exports mirroring ntx-apps/libs/figs/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/figs.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
