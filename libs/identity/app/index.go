package app

import (
	"github.com/thescaffold/gox-apps/libs/identity/pkg"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "identity"

var AllMigrations = Migrations

// AuthModule and AuthMiddleware are re-exported here so host apps only need to
// import this package rather than the pkg sub-package directly.
var AuthModule = &pkg.AuthModule{}
var AuthMiddleware = pkg.AuthMiddleware

// DefaultPermissions mirrors ntx-apps/libs/identity/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {
		"guest:create:apps:identity:initiate:*:*",
		"guest:create:apps:identity:verify:*:*",
		"guest:create:apps:identity:secret:*:*",
		"guest:create:apps:identity:login:*:*",
		"guest:create:apps:identity:logout:*:*",
		"guest:create:apps:identity:reset-secret:initiate:*",
		"guest:create:apps:identity:reset-secret:verify:*",
		"guest:create:apps:identity:reset-secret:update:*",
		"guest:read:apps:identity:provider:*:*",
		"guest:read:apps:identity:provider:providers:*",
		"guest:read:apps:identity:provider:provider:*",
		"guest:create:apps:identity:provider:authorize:*",
		"guest:create:apps:identity:oauth:client:*",
		"guest:read:apps:identity:oauth:client:*",
		"guest:create:apps:identity:oauth:access-token:*",
		"guest:create:apps:identity:oauth:auth-code:*",
		"guest:create:apps:identity:oauth:initiate:*",
		"guest:create:apps:identity:oauth:authorize:*",
		"guest:create:apps:identity:oauth:login:*",
		"guest:read:apps:identity:sessions:*:*",
		"guest:read:apps:identity:now:*:*",
		"guest:create:apps:identity:verify-token:*:*",
		"guest:create:apps:identity:client:*:*",
		"guest:update:apps:identity:client:*:*",
	},
	"member": {
		"member:create:apps:identity:*:*:workspace",
		"member:update:apps:identity:*:*:self",
		"member:read:apps:identity:*:*:workspace",
		"member:delete:apps:identity:*:*:self",
	},
	// Admin scope is `self` per TS index.ts:Admin (not workspace).
	"admin":         {"admin:*:apps:identity:*:*:self"},
	"global-member": {"global-member:*:apps:identity:*:*:global"},
	"global-admin":  {"global-admin:*:apps:identity:*:*:global"},
}

// identityAppSvc is set by AppModule.Declarations() so the subscription
// closures below can dispatch to typed AppService methods after DI.
var identityAppSvc *AppService

// payloadOf extracts the "payload" map from a polylog/eventbus envelope.
func payloadOf(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		if p, ok := m["payload"].(map[string]any); ok {
			return p
		}
		return m
	}
	return nil
}

func decodeEntityCreated(p map[string]any) EntityCreatedPayload {
	e := EntityCreatedPayload{}
	e.EntityID, _ = p["entityId"].(string)
	e.EntityName, _ = p["entityName"].(string)
	if meta, ok := p["meta"].(map[string]any); ok {
		e.UserID, _ = meta["userId"].(string)
		e.ClientID, _ = meta["clientId"].(string)
		e.WorkspaceID, _ = meta["workspaceId"].(string)
		e.Meta = meta
	}
	return e
}

func decodeEntityDeleted(p map[string]any) EntityDeletedPayload {
	d := EntityDeletedPayload{}
	d.EntityID, _ = p["entityId"].(string)
	d.EntityName, _ = p["entityName"].(string)
	if meta, ok := p["meta"].(map[string]any); ok {
		d.UserID, _ = meta["userId"].(string)
		d.ClientID, _ = meta["clientId"].(string)
		d.WorkspaceID, _ = meta["workspaceId"].(string)
		d.Meta = meta
	}
	return d
}

// Subscriptions mirrors TS subscriptions list. Each handler decodes the bus
// payload into a typed struct and forwards to a method on the resolved
// AppService singleton. TS-equivalent bodies are stubbed in AppService.
var Subscriptions = map[string]events.EventHandler{
	"apps.identity.rbac.register": func(_ string, raw any) {
		if identityAppSvc == nil {
			return
		}
		p := payloadOf(raw)
		rp := RBACRegisterPayload{}
		if v, ok := p["roles"].([]any); ok {
			rp.Roles = v
		}
		if v, ok := p["permissions"].([]any); ok {
			rp.Permissions = v
		}
		_ = identityAppSvc.OnRBACRegister(rp)
	},
	"apps.db.identityuserclientworkspaces.after-insert": func(_ string, raw any) {
		if identityAppSvc != nil {
			_ = identityAppSvc.OnUserClientWorkspaceCreated(decodeEntityCreated(payloadOf(raw)))
		}
	},
	"apps.db.identityusers.after-insert": func(_ string, raw any) {
		if identityAppSvc != nil {
			_ = identityAppSvc.OnUserCreated(decodeEntityCreated(payloadOf(raw)))
		}
	},
	"apps.db.identityworkspaces.after-insert": func(_ string, raw any) {
		if identityAppSvc != nil {
			_ = identityAppSvc.OnWorkspaceCreated(decodeEntityCreated(payloadOf(raw)))
		}
	},
	"apps.db.identitydevices.after-insert": func(_ string, raw any) {
		if identityAppSvc != nil {
			_ = identityAppSvc.OnDeviceCreated(decodeEntityCreated(payloadOf(raw)))
		}
	},
	"apps.db.identityworkspaces.after-delete": func(_ string, raw any) {
		if identityAppSvc != nil {
			_ = identityAppSvc.OnWorkspaceDeleted(decodeEntityDeleted(payloadOf(raw)))
		}
	},
	"apps.db.identitydevices.after-delete": func(_ string, raw any) {
		if identityAppSvc != nil {
			_ = identityAppSvc.OnDeviceDeleted(decodeEntityDeleted(payloadOf(raw)))
		}
	},
	"apps.db.identityinvites.after-delete": func(_ string, raw any) {
		if identityAppSvc != nil {
			_ = identityAppSvc.OnInviteDeleted(decodeEntityDeleted(payloadOf(raw)))
		}
	},
	"apps.identity.attribute.update": func(_ string, raw any) {
		if identityAppSvc == nil {
			return
		}
		p := payloadOf(raw)
		ap := AttributeUpdatePayload{}
		ap.UserID, _ = p["userId"].(string)
		ap.Type, _ = p["type"].(string)
		ap.Attributes, _ = p["attributes"].(map[string]any)
		_ = identityAppSvc.OnAttributeUpdate(ap)
	},
	"apps.notification.rule.subscription.update": func(_ string, raw any) {
		if identityAppSvc == nil {
			return
		}
		p := payloadOf(raw)
		sp := SubscriptionUpdatePayload{}
		sp.Ref, _ = p["ref"].(string)
		sp.Rules = p["rules"]
		_ = identityAppSvc.OnSubscriptionUpdate(sp)
	},
}

// Top-level exports mirroring ntx-apps/libs/identity/src/index.ts.
var (
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/identity.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
