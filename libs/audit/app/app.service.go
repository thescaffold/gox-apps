package app

import (
	"encoding/json"
	"strings"

	auditlog "github.com/thescaffold/gox-apps/libs/audit/app/log"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

// UnsafeEntityNames lists entity table names whose events must not be re-logged
// (avoids infinite recursion: logging a log creates another log).
var UnsafeEntityNames = []string{"auditlogs", "audittokens", "audittokenlogs"}

type AppService struct {
	logService *auditlog.LogService   `inject:""`
	tracker    *events.TrackerService `inject:""`
}

func (s *AppService) GetHello() string {
	// TS getHello() returns "Hello World!" uniformly across services.
	return "Hello World!"
}

// HandleEvent creates an audit log entry from an incoming event payload.
// Mirrors the TS handleLogs private method.
func (s *AppService) HandleEvent(payload map[string]any) error {
	entityName, _ := payload["entityName"].(string)
	if entityName == "" {
		return nil
	}
	for _, unsafe := range UnsafeEntityNames {
		if strings.EqualFold(entityName, unsafe) {
			return nil
		}
	}

	meta, _ := payload["meta"].(map[string]any)
	if meta == nil || meta["id"] == nil {
		return nil
	}

	ctx, _ := payload["context"].(map[string]any)
	user, _ := ctx["user"].(map[string]any)
	client, _ := ctx["client"].(map[string]any)
	workspace, _ := ctx["workspace"].(map[string]any)

	if user == nil || client == nil || workspace == nil {
		return nil
	}

	userId, _ := user["id"].(string)
	clientId, _ := client["id"].(string)
	workspaceId, _ := workspace["id"].(string)
	if userId == "" || clientId == "" || workspaceId == "" {
		return nil
	}

	group, _ := payload["group"].(string)
	service, _ := payload["service"].(string)
	action, _ := payload["action"].(string)
	entityId, _ := meta["id"].(string)

	desc := buildDescription(entityName, action, meta)

	delete(meta, "meta") // strip nested sensitive meta

	// Mirror TS app.controller.ts which persists meta + context together.
	combined := map[string]any{}
	for k, v := range meta {
		combined[k] = v
	}
	combined["context"] = ctx
	metaBytes, _ := json.Marshal(combined)

	entry := &auditlog.Log{
		UserId:      userId,
		ClientId:    clientId,
		WorkspaceId: workspaceId,
		Group:       group,
		Service:     service,
		EntityId:    entityId,
		EntityName:  entityName,
		Action:      action,
		Desc:        desc,
		Meta:        metaBytes,
	}

	return s.logService.Create(entry)
}

func (s *AppService) Activities(page, perPage int) ([]auditlog.Log, int64, error) {
	return s.logService.Activities(page, perPage)
}

// buildDescription mirrors the TS descriptions() helper.
func buildDescription(entityName, action string, meta map[string]any) *string {
	if action != "create" && action != "update" && action != "delete" {
		return nil
	}
	var desc string
	switch entityName {
	case "invite":
		ref, _ := meta["ref"].(string)
		typ, _ := meta["type"].(string)
		switch action {
		case "create":
			desc = "Invite sent to " + ref + " via " + typ
		case "update":
			desc = "Invite to " + ref + " via " + typ + " has been updated"
		case "delete":
			desc = "Invite to " + ref + " via " + typ + " has been removed"
		}
	case "workspace":
		name, _ := meta["name"].(string)
		switch action {
		case "create":
			desc = "A new workspace was created: " + name
		case "update":
			desc = "Workspace was updated"
		case "delete":
			desc = "Workspace was removed permanently"
		}
	default:
		desc = titleCase(entityName) + " was " + entityTense(action)
	}
	return &desc
}

func entityTense(action string) string {
	switch action {
	case "create":
		return "created"
	case "update":
		return "updated"
	case "delete":
		return "removed"
	default:
		return action
	}
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
