package app

import (
	"encoding/json"

	identityattribute "github.com/thescaffold/gox-apps-identity/app/attribute"
	identitydevicesession "github.com/thescaffold/gox-apps-identity/app/devicesession"
	identityinvite "github.com/thescaffold/gox-apps-identity/app/invite"
	identitypermission "github.com/thescaffold/gox-apps-identity/app/permission"
	identityrole "github.com/thescaffold/gox-apps-identity/app/role"
	identitytoken "github.com/thescaffold/gox-apps-identity/app/token"
	"github.com/thescaffold/gox-packages-core/events"
)

// AppService is the public surface for identity event handlers.
// Mirrors ntx-apps/libs/identity/src/app.service.ts + the subscription bodies
// in app.controller.ts. Handlers persist directly via injected entities; the
// `bus` is used to re-emit downstream events that other apps subscribe to.
type AppService struct {
	roleEntity          *identityrole.RoleEntity                   `inject:""`
	permissionEntity    *identitypermission.PermissionEntity       `inject:""`
	attributeEntity     *identityattribute.AttributeEntity         `inject:""`
	deviceSessionEntity *identitydevicesession.DeviceSessionEntity `inject:""`
	inviteEntity        *identityinvite.InviteEntity               `inject:""`
	tokenEntity         *identitytoken.TokenEntity                 `inject:""`
	bus                 *events.Bus                                `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

// RBACRegisterPayload is the apps.identity.rbac.register event payload.
type RBACRegisterPayload struct {
	Roles       []any `json:"roles,omitempty"`
	Permissions []any `json:"permissions,omitempty"`
}

// OnRBACRegister persists incoming role + permission definitions, upserting
// each by name (roles) or by key (permissions). Mirrors TS subscription body.
func (s *AppService) OnRBACRegister(p RBACRegisterPayload) error {
	for _, raw := range p.Roles {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		clientID, _ := m["clientId"].(string)
		typ, _ := m["type"].(string)
		if name == "" || clientID == "" {
			continue
		}
		var workspaceID *string
		if v, ok := m["workspaceId"].(string); ok && v != "" {
			workspaceID = &v
		}
		existing, _ := s.roleEntity.First(`client_id = ? AND name = ? AND type = ?`, clientID, name, typ)
		if existing != nil {
			existing.WorkspaceId = workspaceID
			_, _ = s.roleEntity.Update(existing, `id = ?`, existing.Id)
			continue
		}
		_ = s.roleEntity.Insert(&identityrole.Role{
			Name: name, ClientId: clientID, Type: typ, WorkspaceId: workspaceID,
		})
	}

	for _, raw := range p.Permissions {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		key, _ := m["key"].(string)
		scope, _ := m["scope"].(string)
		resource, _ := m["resource"].(string)
		action, _ := m["action"].(string)
		if key == "" {
			continue
		}
		var roleID, clientID, workspaceID *string
		if v, ok := m["roleId"].(string); ok && v != "" {
			roleID = &v
		}
		if v, ok := m["clientId"].(string); ok && v != "" {
			clientID = &v
		}
		if v, ok := m["workspaceId"].(string); ok && v != "" {
			workspaceID = &v
		}
		existing, _ := s.permissionEntity.First(`"key" = ?`, key)
		if existing != nil {
			existing.Scope = scope
			existing.Resource = resource
			existing.Action = action
			existing.RoleId = roleID
			_, _ = s.permissionEntity.Update(existing, `id = ?`, existing.Id)
			continue
		}
		_ = s.permissionEntity.Insert(&identitypermission.Permission{
			Key:         key,
			Scope:       scope,
			Resource:    resource,
			Action:      action,
			RoleId:      roleID,
			ClientId:    clientID,
			WorkspaceId: workspaceID,
		})
	}
	return nil
}

// EntityCreatedPayload covers after-insert events for users/workspaces/devices.
type EntityCreatedPayload struct {
	EntityID    string         `json:"entityId,omitempty"`
	EntityName  string         `json:"entityName,omitempty"`
	UserID      string         `json:"userId,omitempty"`
	ClientID    string         `json:"clientId,omitempty"`
	WorkspaceID string         `json:"workspaceId,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// OnUserCreated handles apps.db.identityusers.after-insert. The TS body is
// intentionally light here — heavy seeding lives in
// OnUserClientWorkspaceCreated which fires after the user-client-workspace row
// is also persisted.
func (s *AppService) OnUserCreated(p EntityCreatedPayload) error { _ = p; return nil }

// OnUserClientWorkspaceCreated chains capital wallet init by re-emitting
// apps.capital.wallet.init on the bus. Mirrors TS subscription body.
func (s *AppService) OnUserClientWorkspaceCreated(p EntityCreatedPayload) error {
	if s.bus == nil {
		return nil
	}
	currency := ""
	if p.Meta != nil {
		if v, ok := p.Meta["currency"].(string); ok {
			currency = v
		}
	}
	s.bus.Publish("apps.capital.wallet.init", map[string]any{
		"payload": map[string]any{
			"userId":      p.UserID,
			"clientId":    p.ClientID,
			"workspaceId": p.WorkspaceID,
			"currency":    currency,
		},
	})
	return nil
}

// OnWorkspaceCreated seeds workspace-scoped admin/member roles for the new
// workspace. Mirrors TS body that inserts default Roles for the workspace.
func (s *AppService) OnWorkspaceCreated(p EntityCreatedPayload) error {
	if p.ClientID == "" {
		return nil
	}
	wid := p.WorkspaceID
	for _, name := range []string{"admin", "member"} {
		existing, _ := s.roleEntity.First(`client_id = ? AND name = ? AND workspace_id = ?`, p.ClientID, name, wid)
		if existing != nil {
			continue
		}
		_ = s.roleEntity.Insert(&identityrole.Role{
			Name: name, ClientId: p.ClientID, WorkspaceId: &wid, Type: "workspace",
		})
	}
	return nil
}

// OnDeviceCreated emits apps.notification.message.new with a "new device"
// alert. Mirrors TS body that pushes a notification.
func (s *AppService) OnDeviceCreated(p EntityCreatedPayload) error {
	if s.bus == nil || p.UserID == "" {
		return nil
	}
	s.bus.Publish("apps.notification.message.new", map[string]any{
		"payload": map[string]any{
			"reference":   "device-" + p.EntityID,
			"userId":      p.UserID,
			"clientId":    p.ClientID,
			"workspaceId": p.WorkspaceID,
			"key":         "new-device",
			"channels":    []string{"email", "web"},
			"subject":     "New device sign-in",
			"data":        map[string]any{"deviceId": p.EntityID, "owner": p.UserID},
			"type":        "security",
			"priority":    "medium",
			"theme":       "warning",
			"scope":       "user",
		},
	})
	return nil
}

// EntityDeletedPayload covers after-delete events.
type EntityDeletedPayload struct {
	EntityID    string         `json:"entityId,omitempty"`
	EntityName  string         `json:"entityName,omitempty"`
	UserID      string         `json:"userId,omitempty"`
	ClientID    string         `json:"clientId,omitempty"`
	WorkspaceID string         `json:"workspaceId,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// OnWorkspaceDeleted soft-deletes workspace-scoped roles/permissions/invites.
func (s *AppService) OnWorkspaceDeleted(p EntityDeletedPayload) error {
	if p.WorkspaceID == "" {
		return nil
	}
	_, _ = s.roleEntity.Delete(`workspace_id = ?`, p.WorkspaceID)
	_, _ = s.permissionEntity.Delete(`workspace_id = ?`, p.WorkspaceID)
	_, _ = s.inviteEntity.Delete(`workspace_id = ?`, p.WorkspaceID)
	return nil
}

// OnDeviceDeleted revokes any active sessions on the device.
func (s *AppService) OnDeviceDeleted(p EntityDeletedPayload) error {
	if p.EntityID == "" {
		return nil
	}
	_, _ = s.deviceSessionEntity.Delete(`device_id = ?`, p.EntityID)
	return nil
}

// OnInviteDeleted removes pending tokens linked to the invite (if the meta
// includes a token reference).
func (s *AppService) OnInviteDeleted(p EntityDeletedPayload) error {
	if p.Meta == nil {
		return nil
	}
	if ref, ok := p.Meta["tokenRef"].(string); ok && ref != "" {
		_, _ = s.tokenEntity.Delete(`token = ?`, ref)
	}
	return nil
}

// AttributeUpdatePayload is the apps.identity.attribute.update event payload.
type AttributeUpdatePayload struct {
	UserID     string         `json:"userId,omitempty"`
	Type       string         `json:"type,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// OnAttributeUpdate persists each attribute key→value pair, then re-emits
// apps.identity.attribute.type.update for downstream subscribers (notification
// listens on this to refresh subscription rules).
func (s *AppService) OnAttributeUpdate(p AttributeUpdatePayload) error {
	if p.UserID == "" {
		return nil
	}
	for key, raw := range p.Attributes {
		val := stringify(raw)
		existing, _ := s.attributeEntity.First(`user_id = ? AND "key" = ?`, p.UserID, key)
		if existing != nil {
			existing.Value = strPtr(val)
			if p.Type != "" {
				existing.Type = strPtr(p.Type)
			}
			_, _ = s.attributeEntity.Update(existing, `id = ?`, existing.Id)
			continue
		}
		_ = s.attributeEntity.Insert(&identityattribute.Attribute{
			UserId: p.UserID, Key: key, Value: strPtr(val), Type: strPtr(p.Type),
		})
	}
	if s.bus != nil {
		ref := p.UserID
		if attr, _ := s.attributeEntity.First(`user_id = ? AND "key" = ?`, p.UserID, "ref"); attr != nil && attr.Value != nil {
			ref = *attr.Value
		}
		s.bus.Publish("apps.identity.attribute.type.update", map[string]any{
			"payload": map[string]any{
				"user":       map[string]any{"id": p.UserID, "ref": ref},
				"type":       p.Type,
				"attributes": p.Attributes,
			},
		})
	}
	return nil
}

// SubscriptionUpdatePayload is the apps.notification.rule.subscription.update payload.
type SubscriptionUpdatePayload struct {
	Ref   string `json:"ref,omitempty"`
	Rules any    `json:"rules,omitempty"`
}

// OnSubscriptionUpdate records the subscription rules as an identity attribute
// keyed by the user's ref. TS body persists this so future identity reads see
// the latest opt-ins.
func (s *AppService) OnSubscriptionUpdate(p SubscriptionUpdatePayload) error {
	if p.Ref == "" {
		return nil
	}
	attr, _ := s.attributeEntity.First(`"key" = ? AND value = ?`, "ref", p.Ref)
	if attr == nil {
		return nil
	}
	rulesType := "notification"
	rulesKey := "subscription"
	rulesValue := stringify(p.Rules)
	existing, _ := s.attributeEntity.First(`user_id = ? AND "key" = ?`, attr.UserId, rulesKey)
	if existing != nil {
		existing.Value = &rulesValue
		existing.Type = &rulesType
		_, _ = s.attributeEntity.Update(existing, `id = ?`, existing.Id)
		return nil
	}
	return s.attributeEntity.Insert(&identityattribute.Attribute{
		UserId: attr.UserId, Key: rulesKey, Value: &rulesValue, Type: &rulesType,
	})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
