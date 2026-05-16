package tagtype

import (
	"regexp"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/common/app/tag"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// TagTypeController mirrors ntx-apps/libs/common/src/api/tag-type/tag-type.controller.ts.
// TS also has morphs.afterGet/afterList that enrich tags via remote
// PlatformService.request (entity-meta map) and uses TrackerService —
// those depend on cross-app entity-meta wiring; tracked for a later
// follow-up. Sync/Attach/Detach are implemented in full here per Phase 7.1.
type TagTypeController struct {
	crud.CrudResource[TagType, CreateTagTypeDto, UpdateTagTypeDto]

	entity     *TagTypeEntity  `inject:""`
	service    *TagTypeService `inject:""`
	tagService *tag.TagService `inject:""`
	lang       *i18n.Service   `inject:""`
}

func (c *TagTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[TagType, CreateTagTypeDto, UpdateTagTypeDto]{
		// Mirrors TS tag-type.controller.ts:29 name = 'tag group'.
		Name: "tag group",
		// Mirrors TS tag-type.controller.ts:30 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS tag-type.controller.ts:32-34 unique = ({workspaceId,name}) => [{...}].
		Unique: func(d *CreateTagTypeDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"name":         d.Name,
			}}
		},
		// Mirrors TS tag-type.controller.ts:70-79 morphs.beforeCreate.
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateTagTypeDto)
				if !ok || p == nil {
					return nil, nil
				}
				if ctx.User != nil {
					if id, ok := ctx.User["id"].(string); ok && id != "" {
						p.UserId = &id
					}
				}
				if ctx.Client != nil {
					if id, ok := ctx.Client["id"].(string); ok && id != "" {
						p.ClientId = &id
					}
				}
				if ctx.Workspace != nil {
					if id, ok := ctx.Workspace["id"].(string); ok && id != "" {
						p.WorkspaceId = &id
					}
				}
				return p, nil
			},
		},
	})
	c.SetLang(c.lang)
}

// uuidPattern matches TS' /^[0-9a-f]{8}-[0-9a-f]{4}-…/ — used to detect when
// a name in the sync request is actually a tag-type id that should be looked
// up first.
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Sync handles POST /tag-type/sync — diffs the supplied names against the
// existing tag-types linked to the entity, detaches the removed ones,
// attaches the new ones, leaves the unchanged ones alone. Mirrors TS
// tag-type.controller.ts:122-175 syncTag().
func (c *TagTypeController) Sync(dto *TagActionDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.common.title", nil, pref)

	resolvedNames := make([]string, 0, len(dto.Names))
	for _, name := range dto.Names {
		if uuidPattern.MatchString(name) {
			if tt, _ := c.service.FindOne(map[string]any{"id": name}); tt != nil && tt.Name != "" {
				resolvedNames = append(resolvedNames, tt.Name)
				continue
			}
		}
		resolvedNames = append(resolvedNames, name)
	}

	existingNames := c.namesLinkedToEntity(dto)
	onlyInExisting, onlyInNew, _ := diff(existingNames, resolvedNames)

	if out := c.detachNames(dto, onlyInExisting); out != nil {
		return out
	}
	c.attachNames(dto, onlyInNew)

	return response.Success(nil,
		title,
		c.lang.Translate("apps.common.tag-type.post.sync.success", nil, pref),
		nil)
}

// Attach handles POST /tag-type/attach — find-or-create each named tag-type
// and link it to the entity. Mirrors TS attachTag().
func (c *TagTypeController) Attach(dto *TagActionDto) types.Output {
	pref := dto.Ctx.Preference
	c.attachNames(dto, dto.Names)
	return response.Success(nil,
		c.lang.Translate("apps.common.title", nil, pref),
		c.lang.Translate("apps.common.tag-type.post.attach.success", nil, pref),
		nil)
}

// Detach handles POST /tag-type/detach — soft-delete each named link; when
// the tag-type has no remaining tags, soft-delete the tag-type itself.
// Mirrors TS detachTag().
func (c *TagTypeController) Detach(dto *TagActionDto) types.Output {
	pref := dto.Ctx.Preference
	if out := c.detachNames(dto, dto.Names); out != nil {
		return out
	}
	return response.Success(nil,
		c.lang.Translate("apps.common.title", nil, pref),
		c.lang.Translate("apps.common.tag-type.post.detach.success", nil, pref),
		nil)
}

// namesLinkedToEntity returns the distinct tag-type names that currently have
// a tag row pointing at (group, service, entity, entityId). Mirrors TS:
//
//	tagTypeService.find({where:[{tags:{groupName,serviceName,entityName,entityId}}]}).map(t=>t.name)
//
// The relational TypeORM filter is collapsed in gox to a SQL join — we read
// the tag rows directly, then look up the type names by id.
func (c *TagTypeController) namesLinkedToEntity(dto *TagActionDto) []string {
	if c.tagService == nil {
		return nil
	}
	rows, err := c.tagService.Entity().Some(
		`"group_name" = ? AND "service_name" = ? AND "entity_name" = ? AND "entity_id" = ?`,
		dto.GroupName, dto.ServiceName, dto.EntityName, dto.EntityId,
	)
	if err != nil || len(rows) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.TypeId == "" {
			continue
		}
		tt, err := c.service.FindOne(map[string]any{"id": r.TypeId})
		if err != nil || tt == nil || tt.Name == "" {
			continue
		}
		if _, ok := seen[tt.Name]; ok {
			continue
		}
		seen[tt.Name] = struct{}{}
		out = append(out, tt.Name)
	}
	return out
}

// attachNames find-or-creates a TagType per supplied name and links it to the
// entity via Tag.UpdateOrCreate. Mirrors TS attachNames() — derives the owner
// (user/client/workspace) from request context.
func (c *TagTypeController) attachNames(dto *TagActionDto, names []string) {
	if len(names) == 0 || c.service == nil || c.tagService == nil {
		return
	}
	userID, clientID, workspaceID := ownerFromCtx(dto.Ctx)
	for _, name := range names {
		match := map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"name":         name,
		}
		tt, err := c.service.FindOrCreate(match)
		if err != nil || tt == nil {
			continue
		}
		_, _ = c.tagService.UpdateOrCreate(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"group_name":   dto.GroupName,
			"service_name": dto.ServiceName,
			"entity_name":  dto.EntityName,
			"entity_id":    dto.EntityId,
			"type_id":      tt.Id,
		})
	}
}

// detachNames soft-deletes the link rows; when a TagType ends up with no
// remaining tags it gets soft-deleted too. Mirrors TS detachNames().
// Returns a non-nil error envelope when one of the named tag-types cannot
// be located.
func (c *TagTypeController) detachNames(dto *TagActionDto, names []string) types.Output {
	if len(names) == 0 || c.service == nil || c.tagService == nil {
		return nil
	}
	pref := dto.Ctx.Preference
	userID, clientID, workspaceID := ownerFromCtx(dto.Ctx)
	for _, name := range names {
		tt, err := c.service.FindOne(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"name":         name,
		})
		if err != nil || tt == nil {
			return response.BadRequest(
				c.lang.Translate("apps.common.title", nil, pref),
				c.lang.Translate("apps.common.tag-type.error.tag-not-found", nil, pref),
			)
		}
		_, _ = c.tagService.SoftDelete(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"group_name":   dto.GroupName,
			"service_name": dto.ServiceName,
			"entity_name":  dto.EntityName,
			"entity_id":    dto.EntityId,
			"type_id":      tt.Id,
		})
		// If the type now has zero remaining links, retire it too.
		if remaining, _ := c.tagService.Count(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"type_id":      tt.Id,
		}); remaining < 1 {
			_, _ = c.service.SoftDelete(map[string]any{"id": tt.Id})
		}
	}
	return nil
}

// diff splits A vs B into (onlyInA, onlyInB, inBoth). Mirrors TS compareTagTypeNames.
func diff(a, b []string) (onlyInA, onlyInB, inBoth []string) {
	inB := map[string]bool{}
	for _, v := range b {
		inB[v] = true
	}
	inA := map[string]bool{}
	for _, v := range a {
		inA[v] = true
	}
	for _, v := range a {
		if inB[v] {
			inBoth = append(inBoth, v)
		} else {
			onlyInA = append(onlyInA, v)
		}
	}
	for _, v := range b {
		if !inA[v] {
			onlyInB = append(onlyInB, v)
		}
	}
	return
}

// ownerFromCtx extracts the (user, client, workspace) trio from the request
// context, falling back to the empty string when a header is absent.
func ownerFromCtx(ctx ntxctx.NTXContext) (string, string, string) {
	var u, c, w string
	if ctx.User != nil {
		u, _ = ctx.User["id"].(string)
	}
	if ctx.Client != nil {
		c, _ = ctx.Client["id"].(string)
	}
	if ctx.Workspace != nil {
		w, _ = ctx.Workspace["id"].(string)
	}
	return u, c, w
}
