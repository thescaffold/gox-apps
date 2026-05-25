package voucher

import (
	"encoding/json"
	"time"

	"github.com/awesome-goose/goose/types"
	capitalvouchertype "github.com/thescaffold/gox-apps/libs/capital/app/vouchertype"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// VoucherController mirrors ntx-apps/libs/capital/src/api/voucher/voucher.controller.ts.
// Implements:
//   - morphs.beforeCreate: hydrate user/client/workspace from context
//   - morphs.afterGet: enrich the row with VoucherType, stripped of `rules`
//   - morphs.afterList: same enrichment across the page
type VoucherController struct {
	crud.CrudResource[Voucher, CreateVoucherDto, UpdateVoucherDto]

	entity            *VoucherEntity                            `inject:""`
	voucherTypeEntity *capitalvouchertype.VoucherTypeEntity     `inject:""`
	lang              *i18n.Service                             `inject:""`
}

func (c *VoucherController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Voucher, CreateVoucherDto, UpdateVoucherDto]{
		Name:       "voucher",
		Searchable: []string{},
		Unique: func(d *CreateVoucherDto) []utils.KeyValue {
			return []utils.KeyValue{{
				"workspace_id":    d.WorkspaceId,
				"voucher_type_id": d.VoucherTypeId,
			}}
		},
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateVoucherDto)
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
						p.WorkspaceId = id
					}
				}
				return p, nil
			},
			crud.AfterGet: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				v, ok := payload.(*Voucher)
				if !ok || v == nil {
					return payload, nil
				}
				return enrichVoucher(c, v), nil
			},
			crud.AfterList: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				rows, ok := payload.([]Voucher)
				if !ok {
					return payload, nil
				}
				out := make([]map[string]any, 0, len(rows))
				for i := range rows {
					out = append(out, enrichVoucher(c, &rows[i]))
				}
				return out, nil
			},
		},
	})
	c.SetLang(c.lang)
}

// ListAvailable mirrors TS @Get('list') — iterates every VoucherType and
// returns only those whose `rules` allow the current (user, client, workspace,
// scope) tuple. The `rules` field is stripped from response payloads.
func (c *VoucherController) ListAvailable(dto *VoucherListDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.capital.voucher.title", nil, pref)
	if c.voucherTypeEntity == nil {
		return response.Success([]any{}, title,
			c.lang.Translate("apps.capital.voucher.success", nil, pref), nil)
	}
	total, _ := c.voucherTypeEntity.Count(``)
	perPage := int64(100)
	pages := (total + perPage - 1) / perPage
	available := []any{}
	for page := int64(1); page <= pages; page++ {
		types, _ := c.voucherTypeEntity.Find(int((page-1)*perPage), int(perPage), ``)
		for i := range types {
			vt := &types[i]
			if !c.isAvailable(dto.Ctx, dto.Scope, vt) {
				continue
			}
			stripped := map[string]any{
				"id":       vt.Id,
				"name":     vt.Name,
				"desc":     vt.Desc,
				"token":    vt.Token,
				"amount":   vt.Amount,
				"currency": vt.Currency,
				"status":   vt.Status,
			}
			available = append(available, stripped)
		}
	}
	return response.Success(available, title,
		c.lang.Translate("apps.capital.voucher.success", nil, pref), nil)
}

// Verify mirrors TS @Post('verify/:token') — looks up the VoucherType by
// token, runs isAvailable, and persists a Voucher row for the (user, client,
// workspace, voucherTypeId) tuple on success.
func (c *VoucherController) Verify(dto *VoucherVerifyDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.capital.voucher.title", nil, pref)
	if c.voucherTypeEntity == nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.capital.voucher.error.invalid-voucher-type", nil, pref))
	}
	vt, _ := c.voucherTypeEntity.First(`"token" = ?`, dto.Token)
	if vt == nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.capital.voucher.error.invalid-voucher-type", nil, pref))
	}
	if !c.isAvailable(dto.Ctx, dto.Scope, vt) {
		return response.BadRequest(title,
			c.lang.Translate("apps.capital.voucher.error.invalid-voucher-type", nil, pref))
	}
	userId, clientId, workspaceId := voucherTriple(dto.Ctx)
	row := &Voucher{
		UserId:        userId,
		ClientId:      clientId,
		WorkspaceId:   workspaceId,
		VoucherTypeId: vt.Id,
	}
	_ = c.entity.Insert(row)
	return response.Success(vt, title,
		c.lang.Translate("apps.capital.voucher.success", nil, pref), nil)
}

// isAvailable runs the TS-equivalent rule checks on a VoucherType.rules
// payload: reuse-count cap, time window, allowed emails/clients/scopes.
func (c *VoucherController) isAvailable(ntx ntxctx.NTXContext, scope string, vt *capitalvouchertype.VoucherType) bool {
	if vt == nil {
		return false
	}
	var rules map[string]any
	if len(vt.Rules) > 0 {
		_ = json.Unmarshal(vt.Rules, &rules)
	}
	// TS does NOT short-circuit when rules is absent: `reuse = rules?.reuse ?? 0`,
	// so a voucher type already used at least once is unavailable even with no
	// rules. Go map reads below are nil-safe, so run the checks with a nil ruleset.

	// reuse cap
	reuseCap := 0
	if v, ok := rules["reuse"].(float64); ok {
		reuseCap = int(v)
	}
	used, _ := c.entity.Count(`"voucher_type_id" = ?`, vt.Id)
	if reuseCap < int(used) {
		return false
	}

	now := time.Now().UTC()
	if startAt, ok := rules["startAt"].(string); ok && startAt != "" {
		if t, err := time.Parse(time.RFC3339, startAt); err == nil && t.After(now) {
			return false
		}
	}
	if endAt, ok := rules["endAt"].(string); ok && endAt != "" {
		if t, err := time.Parse(time.RFC3339, endAt); err == nil && t.Before(now) {
			return false
		}
	}

	// userRef gate
	if emails, ok := rules["emails"].([]any); ok && len(emails) > 0 {
		userRef := ""
		if ntx.User != nil {
			userRef, _ = ntx.User["ref"].(string)
			if userRef == "" {
				userRef, _ = ntx.User["email"].(string)
			}
		}
		if !containsString(emails, userRef) {
			return false
		}
	}
	if clients, ok := rules["clients"].([]any); ok && len(clients) > 0 {
		clientName := ""
		if ntx.Client != nil {
			clientName, _ = ntx.Client["name"].(string)
		}
		if !containsString(clients, clientName) {
			return false
		}
	}
	if scopes, ok := rules["scopes"].([]any); ok && len(scopes) > 0 {
		if !containsString(scopes, scope) {
			return false
		}
	}
	return true
}

func containsString(list []any, target string) bool {
	for _, v := range list {
		if s, ok := v.(string); ok && s == target {
			return true
		}
	}
	return false
}

func voucherTriple(ntx ntxctx.NTXContext) (string, string, string) {
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

// enrichVoucher returns a Voucher + VoucherType envelope with the type's
// `rules` field stripped (TS hides rules from response payloads).
func enrichVoucher(c *VoucherController, v *Voucher) map[string]any {
	envelope := map[string]any{
		"id":            v.Id,
		"userId":        v.UserId,
		"clientId":      v.ClientId,
		"workspaceId":   v.WorkspaceId,
		"voucherTypeId": v.VoucherTypeId,
		"status":        v.Status,
		"createdAt":     v.CreatedAt,
		"updatedAt":     v.UpdatedAt,
	}
	if c.voucherTypeEntity != nil {
		if vt, _ := c.voucherTypeEntity.First(`"id" = ?`, v.VoucherTypeId); vt != nil {
			envelope["voucherType"] = map[string]any{
				"id":       vt.Id,
				"name":     vt.Name,
				"desc":     vt.Desc,
				"token":    vt.Token,
				"amount":   vt.Amount,
				"currency": vt.Currency,
				"status":   vt.Status,
			}
		}
	}
	return envelope
}
