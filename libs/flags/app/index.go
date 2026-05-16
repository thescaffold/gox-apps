package app

import (
	"encoding/json"
	"os"

	flagsenvironment "github.com/thescaffold/gox-apps/libs/flags/app/environment"
	flagsenvironmenttype "github.com/thescaffold/gox-apps/libs/flags/app/environmenttype"
	flagsflag "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	flagsflaglog "github.com/thescaffold/gox-apps/libs/flags/app/flaglog"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "flags"

var AllMigrations = Migrations

// DefaultPermissions mirrors ntx-apps/libs/flags/src/index.ts defaultRolesNPermissions.
var DefaultPermissions = map[string][]string{
	"guest": {
		"guest:read:apps:flags:remote:*:self",
		"guest:read:apps:flags:dynamic:*:self",
		"guest:create:apps:flags:file:*:self",
		"guest:update:apps:flags:file:*:self",
		"guest:read:apps:flags:file:*:self",
	},
	"member": {
		"member:create:apps:flags:*:*:workspace",
		"member:update:apps:flags:*:*:self",
		"member:read:apps:flags:*:*:workspace",
		"member:delete:apps:flags:*:*:self",
	},
	"admin":         {"admin:*:apps:flags:*:*:workspace"},
	"global-member": {"global-member:*:apps:flags:*:*:global"},
	"global-admin":  {"global-admin:*:apps:flags:*:*:global"},
}

// flagsAppSvc holds the DI-managed AppService so the top-level subscription
// handlers below can drive it without participating in struct injection.
// AppModule.Declarations() assigns this before the kernel boots.
var flagsAppSvc *AppService

// payloadMap drills into `item.payload` — TS subscriptions receive `Item`s
// whose `payload` field carries the actual event data.
func payloadMap(raw any) map[string]any {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	if inner, ok := m["payload"].(map[string]any); ok {
		return inner
	}
	return m
}

// envRef extracts an EnvironmentRef-shaped object from a generic map entry.
// Returns the zero value when the key is missing/wrong-typed.
func envRef(m map[string]any, key string) EnvironmentRef {
	v, ok := m[key].(map[string]any)
	if !ok {
		return EnvironmentRef{}
	}
	out := EnvironmentRef{}
	if name, ok := v["name"].(string); ok {
		out.Name = name
	}
	if id, ok := v["id"].(string); ok && id != "" {
		out.ID = &id
	}
	return out
}

// handleFlagRegister mirrors TS subscription 'apps.flags.flag.register':
// defaults missing environmentType to {name:'Javascript'} and environment to
// {name: SOURCE_ID}, then calls appService.register with the supplied flags.
func handleFlagRegister(_ string, payload any) {
	if flagsAppSvc == nil {
		return
	}
	body := payloadMap(payload)
	if body == nil {
		return
	}
	envType := envRef(body, "environmentType")
	if envType.Name == "" && envType.ID == nil {
		envType = EnvironmentRef{Name: "Javascript"}
	}
	env := envRef(body, "environment")
	if env.Name == "" && env.ID == nil {
		env = EnvironmentRef{Name: os.Getenv("SOURCE_ID")}
	}
	rawFlags, _ := body["flags"].([]any)
	flags := make([]FlagDef, 0, len(rawFlags))
	owner := FlagOwner{}
	for _, rf := range rawFlags {
		fm, ok := rf.(map[string]any)
		if !ok {
			continue
		}
		def := FlagDef{}
		if name, ok := fm["name"].(string); ok {
			def.Name = name
		}
		if limit, ok := fm["limit"].(float64); ok {
			def.Limit = int(limit)
		}
		if priority, ok := fm["priority"].(float64); ok {
			def.Priority = int(priority)
		}
		if level, ok := fm["level"].(string); ok {
			def.Level = level
		}
		if status, ok := fm["status"].(string); ok {
			def.Status = &status
		}
		if metaRaw, ok := fm["meta"]; ok && metaRaw != nil {
			if b, err := json.Marshal(metaRaw); err == nil {
				def.Meta = b
			}
		}
		// owner is per-flag in the TS handler; if any flag carries its own
		// userId/clientId/workspaceId, it shadows the (otherwise zero) owner.
		if uid, ok := fm["userId"].(string); ok && uid != "" {
			owner.UserId = uid
		}
		if cid, ok := fm["clientId"].(string); ok && cid != "" {
			owner.ClientId = cid
		}
		if wid, ok := fm["workspaceId"].(string); ok && wid != "" {
			owner.WorkspaceId = wid
		}
		flags = append(flags, def)
	}
	_, _ = flagsAppSvc.Register(envType, env, owner, flags)
}

// handleFlagLog mirrors TS subscription 'apps.flags.flag.log': resolves a
// flag tuple from payload.flag, defaults missing environment to {name: SOURCE_ID},
// then writes a FlagLog with the supplied limit (default 1).
func handleFlagLog(_ string, payload any) {
	if flagsAppSvc == nil {
		return
	}
	body := payloadMap(payload)
	if body == nil {
		return
	}
	env := envRef(body, "environment")
	if env.Name == "" && env.ID == nil {
		env = EnvironmentRef{Name: os.Getenv("SOURCE_ID")}
	}
	flagMap, _ := body["flag"].(map[string]any)
	if flagMap == nil {
		return
	}
	q := LogQuery{}
	if v, ok := flagMap["userId"].(string); ok {
		q.UserId = v
	}
	if v, ok := flagMap["clientId"].(string); ok {
		q.ClientId = v
	}
	if v, ok := flagMap["workspaceId"].(string); ok {
		q.WorkspaceId = v
	}
	if v, ok := flagMap["name"].(string); ok {
		q.Name = v
	}
	if v, ok := flagMap["level"].(string); ok {
		q.Level = v
	}
	limit := 1
	if v, ok := body["limit"].(float64); ok && int(v) > 0 {
		limit = int(v)
	}
	_, _ = flagsAppSvc.Log(env, q, limit)
}

// Subscriptions mirrors ntx-apps/libs/flags/src/index.ts subscriptions exactly.
var Subscriptions = map[string]events.EventHandler{
	"apps.flags.flag.register": handleFlagRegister,
	"apps.flags.flag.log":      handleFlagLog,
}

// Top-level exports mirroring ntx-apps/libs/flags/src/index.ts.
var (
	Entities = []any{
		flagsflag.Flag{},
		flagsenvironment.Environment{},
		flagsenvironmenttype.EnvironmentType{},
		flagsflaglog.FlagLog{},
	}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/flags.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
