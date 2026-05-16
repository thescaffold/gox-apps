package token

import (
	"time"

	"github.com/awesome-goose/goose/types"
	identityrole "github.com/thescaffold/gox-apps/libs/identity/app/role"
	identityroletype "github.com/thescaffold/gox-apps/libs/identity/app/roletype"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

const (
	defaultDurationSec = 100 * 365 * 24 * 60 * 60 // 100y in seconds
	tokenValueLen      = 32
	tokenCheckLen      = 16
)

// TokenController mirrors ntx-apps/libs/identity/src/api/token/token.controller.ts.
// Ported behaviour:
//   - morphs.beforeCreate: hydrate userId/clientId/workspaceId from context +
//     mint a fresh 32-char value + 16-char check + expiredAt window.
//   - morphsCtx.afterCreate: attach `maskedValue` to the response envelope.
//   - hooksCtx.afterCreate: when status='platform', look up Admin role-type,
//     decide global-admin vs. global-member, then save Role rows binding the
//     token id to each platform RoleType.
//   - POST /token/user: find-or-create a Token for the (user,client,workspace,
//     name) tuple, mint fresh value/check on ?refresh, seed Role rows for
//     each context role.
type TokenController struct {
	crud.CrudResource[Token, CreateTokenDto, UpdateTokenDto]

	entity         *TokenEntity                     `inject:""`
	roleEntity     *identityrole.RoleEntity         `inject:""`
	roleTypeEntity *identityroletype.RoleTypeEntity `inject:""`
	lang           *i18n.Service                    `inject:""`
}

func (c *TokenController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Token, CreateTokenDto, UpdateTokenDto]{
		Name:       "token",
		Searchable: []string{"name", "desc"},
		Unique: func(d *CreateTokenDto) []utils.KeyValue {
			out := utils.KeyValue{}
			if d.Name != nil {
				out["name"] = *d.Name
			} else {
				out["type"] = d.Type
			}
			if d.WorkspaceId != nil {
				out["workspace_id"] = *d.WorkspaceId
			} else {
				out["workspace_id"] = nil
			}
			return []utils.KeyValue{out}
		},
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateTokenDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = &id
					}
				}
				duration := defaultDurationSec
				if p.DurationInSec != nil && *p.DurationInSec > 0 {
					duration = *p.DurationInSec
				}
				val := utils.Random(tokenValueLen)
				chk := utils.Random(tokenCheckLen)
				exp := time.Now().UTC().Add(time.Duration(duration) * time.Second)
				p.Value = &val
				p.Check = &chk
				p.ExpiredAt = &exp
				return p, nil
			},
		},
		MorphsCtx: map[string]crud.MorphCtxFn{
			crud.AfterCreate: func(event crud.HookEvent, ctx ntxctx.NTXContext) (any, error) {
				t, ok := event.Entity.(*Token)
				if !ok || t == nil {
					return event.Entity, nil
				}
				if t.Value != nil {
					mv := utils.Mask(*t.Value)
					t.MaskedValue = &mv
				}
				return t, nil
			},
		},
		HooksCtx: map[string]crud.HookCtxFn{
			crud.AfterCreate: func(event crud.HookEvent, ctx ntxctx.NTXContext) error {
				t, _ := event.Entity.(*Token)
				dto, _ := event.DTO.(*CreateTokenDto)
				if t == nil || dto == nil || c.roleEntity == nil || c.roleTypeEntity == nil {
					return nil
				}
				if t.Status == nil || *t.Status != "platform" {
					return nil
				}
				isAdmin := false
				for _, rtId := range dto.RoleTypeIds {
					rt, _ := c.roleTypeEntity.First(`"id" = ? AND "name" = ?`, rtId, "admin")
					if rt != nil {
						isAdmin = true
						break
					}
				}
				platformRoleName := "global-member"
				if isAdmin {
					platformRoleName = "global-admin"
				}
				platformRoleTypes, _ := c.roleTypeEntity.Find(0, 0,
					`"name" = ?`, platformRoleName)
				ids := append([]string{}, dto.RoleTypeIds...)
				for _, rt := range platformRoleTypes {
					ids = append(ids, rt.Id)
				}
				workspaceId := ""
				if t.WorkspaceId != nil {
					workspaceId = *t.WorkspaceId
				}
				entityName := "token"
				for _, rtId := range ids {
					rtCopy := rtId
					tCopy := t.Id
					_ = c.roleEntity.Insert(&identityrole.Role{
						Name:        rtId,
						ClientId:    t.ClientId,
						WorkspaceId: &workspaceId,
						Type:        entityName,
						EntityName:  &entityName,
						EntityId:    &tCopy,
						RoleTypeId:  &rtCopy,
					})
				}
				return nil
			},
		},
	})
	c.SetLang(c.lang)
}

// User mirrors TS @Post('user') — find-or-create a Token for the
// (userId, clientId, workspaceId, name) tuple, mint fresh value/check on
// ?refresh, and seed platform Role rows for each context role.
func (c *TokenController) User(dto *UserTokenDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.token.title", nil, pref)
	userId, clientId, workspaceId := tokenTriple(dto.Ctx)
	if userId == "" {
		return response.Unauthorized(title, "authentication required")
	}
	existing, _ := c.entity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "name" = ?`,
		userId, clientId, workspaceId, dto.Name,
	)
	refresh := dto.Refresh != ""
	if existing == nil || refresh {
		duration := defaultDurationSec
		if dto.DurationInSec != nil && *dto.DurationInSec > 0 {
			duration = *dto.DurationInSec
		}
		value := utils.Random(tokenValueLen)
		check := utils.Random(tokenCheckLen)
		exp := time.Now().UTC().Add(time.Duration(duration) * time.Second)
		name := dto.Name
		ws := workspaceId
		if existing != nil {
			existing.Desc = dto.Desc
			existing.Value = &value
			existing.Check = &check
			existing.ExpiredAt = &exp
			_, _ = c.entity.Update(existing, `"id" = ?`, existing.Id)
		} else {
			existing = &Token{
				UserId: userId, ClientId: clientId, WorkspaceId: &ws,
				Type: "user", Token: value,
				Name: &name, Desc: dto.Desc, Value: &value, Check: &check,
				ExpiredAt: &exp,
			}
			_ = c.entity.Insert(existing)
		}
		// Seed platform-role join rows for each ctx-resolved role.
		if c.roleEntity != nil && c.roleTypeEntity != nil {
			for _, role := range rolesFromContext(dto.Ctx) {
				rt, _ := c.roleTypeEntity.First(`"name" = ?`, role)
				if rt == nil {
					continue
				}
				rtId := rt.Id
				tCopy := existing.Id
				entityName := "token"
				_ = c.roleEntity.Insert(&identityrole.Role{
					Name:        rt.Name,
					ClientId:    clientId,
					WorkspaceId: &ws,
					Type:        entityName,
					EntityName:  &entityName,
					EntityId:    &tCopy,
					RoleTypeId:  &rtId,
				})
			}
		}
	}
	if existing.Value != nil {
		mv := utils.Mask(*existing.Value)
		existing.MaskedValue = &mv
	}
	return response.Success(existing, title,
		c.lang.Translate("apps.identity.token.success", nil, pref), nil)
}

func tokenTriple(ntx ntxctx.NTXContext) (string, string, string) {
	userId, clientId, workspaceId := "", "", ""
	if ntx.User != nil {
		userId, _ = ntx.User["id"].(string)
	}
	if userId == "" {
		userId = ntx.UserID
	}
	if ntx.Client != nil {
		clientId, _ = ntx.Client["id"].(string)
	}
	if clientId == "" {
		clientId = ntx.ClientID
	}
	if ntx.Workspace != nil {
		workspaceId, _ = ntx.Workspace["id"].(string)
	}
	if workspaceId == "" {
		workspaceId = ntx.WorkspaceID
	}
	return userId, clientId, workspaceId
}

func rolesFromContext(ntx ntxctx.NTXContext) []string {
	if ntx.Roles == nil {
		return nil
	}
	out := make([]string, 0, len(ntx.Roles))
	for _, r := range ntx.Roles {
		if s, ok := r.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}
