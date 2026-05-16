package projecttype

import (
	"os"
	"regexp"

	"github.com/awesome-goose/goose/types"
	assetsapp "github.com/thescaffold/gox-apps/libs/assets/app"
	"github.com/thescaffold/gox-apps/libs/common/app/project"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// ProjectTypeController mirrors ntx-apps/libs/common/src/api/project-type/project-type.controller.ts.
// TS also has hooks.beforeList that seeds a 'Default' project type +
// morphs.afterGet/afterList that enrich projects with remote entity data via
// PlatformService — those are tracked as follow-ups; the BeforeCreate
// dynamic-asset thumbnail branch lands separately (Phase 7.1 thumbnail step).
// Sync/Attach/Detach are implemented in full here (Phase 7.1).
type ProjectTypeController struct {
	crud.CrudResource[ProjectType, CreateProjectTypeDto, UpdateProjectTypeDto]

	entity         *ProjectTypeEntity      `inject:""`
	service        *ProjectTypeService     `inject:""`
	projectService *project.ProjectService `inject:""`
	assetsApp      *assetsapp.AppService   `inject:""`
	lang           *i18n.Service           `inject:""`
}

func (c *ProjectTypeController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[ProjectType, CreateProjectTypeDto, UpdateProjectTypeDto]{
		// Mirrors TS project-type.controller.ts:30 name = 'project group'.
		Name: "project group",
		// Mirrors TS project-type.controller.ts:31 searchable = ['name','desc'].
		Searchable: []string{"name", "desc"},
		// Mirrors TS project-type.controller.ts:33-35 unique = ({workspaceId,name}) => [{...}].
		Unique: func(d *CreateProjectTypeDto) []map[string]any {
			return []map[string]any{{
				"workspace_id": d.WorkspaceId,
				"name":         d.Name,
			}}
		},
		Morphs: map[string]crud.MorphFn{
			crud.BeforeCreate: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				p, ok := payload.(*CreateProjectTypeDto)
				if !ok || p == nil {
					return nil, nil
				}
				// TS project-type.controller.ts:77-101 BeforeCreate: when the
				// payload has no thumbnailUrl, request a freshly-generated
				// dynamic asset from assets.AppService and assign its URL.
				if p.ThumbnailUrl == nil || *p.ThumbnailUrl == "" {
					if url := c.generateThumbnail(); url != "" {
						p.ThumbnailUrl = &url
					}
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

// uuidPattern mirrors TS' /^[0-9a-f]{8}-[0-9a-f]{4}-…/ — used to detect when
// a name in the sync request is actually a project-type id that should be
// looked up first.
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Sync handles POST /project-type/sync — diffs the supplied names against
// the existing project-types linked to the entity, detaches removed ones,
// attaches new ones, leaves unchanged ones alone. Mirrors TS
// project-type.controller.ts:172-231 syncProject().
func (c *ProjectTypeController) Sync(dto *ProjectActionDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.common.title", nil, pref)

	resolvedNames := make([]string, 0, len(dto.Names))
	for _, name := range dto.Names {
		if uuidPattern.MatchString(name) {
			if pt, _ := c.service.FindOne(map[string]any{"id": name}); pt != nil && pt.Name != "" {
				resolvedNames = append(resolvedNames, pt.Name)
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
		c.lang.Translate("apps.common.project-type.post.sync.success", nil, pref),
		nil)
}

// Attach handles POST /project-type/attach.
func (c *ProjectTypeController) Attach(dto *ProjectActionDto) types.Output {
	pref := dto.Ctx.Preference
	c.attachNames(dto, dto.Names)
	return response.Success(nil,
		c.lang.Translate("apps.common.title", nil, pref),
		c.lang.Translate("apps.common.project-type.post.attach.success", nil, pref),
		nil)
}

// Detach handles POST /project-type/detach.
func (c *ProjectTypeController) Detach(dto *ProjectActionDto) types.Output {
	pref := dto.Ctx.Preference
	if out := c.detachNames(dto, dto.Names); out != nil {
		return out
	}
	return response.Success(nil,
		c.lang.Translate("apps.common.title", nil, pref),
		c.lang.Translate("apps.common.project-type.post.detach.success", nil, pref),
		nil)
}

// namesLinkedToEntity returns the distinct project-type names that currently
// have a project row pointing at (group, service, entity, entityId). Mirrors
// TS's relational filter.
func (c *ProjectTypeController) namesLinkedToEntity(dto *ProjectActionDto) []string {
	if c.projectService == nil {
		return nil
	}
	rows, err := c.projectService.Entity().Some(
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
		pt, err := c.service.FindOne(map[string]any{"id": r.TypeId})
		if err != nil || pt == nil || pt.Name == "" {
			continue
		}
		if _, ok := seen[pt.Name]; ok {
			continue
		}
		seen[pt.Name] = struct{}{}
		out = append(out, pt.Name)
	}
	return out
}

// attachNames find-or-creates a ProjectType per supplied name and links it
// to the entity via Project.UpdateOrCreate. Mirrors TS attachNames() —
// includes the dynamic-asset thumbnail (TS calls assetsAppService.getDynamicAsset
// per new project-type and stores the resulting URL as thumbnailUrl).
func (c *ProjectTypeController) attachNames(dto *ProjectActionDto, names []string) {
	if len(names) == 0 || c.service == nil || c.projectService == nil {
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
		var defaults map[string]any
		if url := c.generateThumbnail(); url != "" {
			defaults = map[string]any{"thumbnail_url": url}
		}
		pt, err := c.service.FindOrCreate(match, defaults)
		if err != nil || pt == nil {
			continue
		}
		_, _ = c.projectService.UpdateOrCreate(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"group_name":   dto.GroupName,
			"service_name": dto.ServiceName,
			"entity_name":  dto.EntityName,
			"entity_id":    dto.EntityId,
			"type_id":      pt.Id,
		})
	}
}

// generateThumbnail mirrors TS `assetsAppService.getDynamicAsset()` — it
// asks the assets app to generate + store a fresh dynamic SVG and returns
// the served URL. Returns "" on any failure so the caller can decide whether
// to proceed without a thumbnail (TS errors out; gox is more permissive to
// keep the sync flow moving — the row still gets created, just without a
// thumbnailUrl).
func (c *ProjectTypeController) generateThumbnail() string {
	if c.assetsApp == nil {
		return ""
	}
	name := assetsapp.RandomName()
	svg := c.assetsApp.GenerateDynamicSVG(name, "pixel", nil, 0, false)
	baseURL := os.Getenv("ASSETS_BASE_URL")
	_, url, err := c.assetsApp.StoreDynamicFile(name, svg, baseURL)
	if err != nil {
		return ""
	}
	return url
}

// detachNames soft-deletes the link rows; when a ProjectType has no remaining
// projects it gets soft-deleted too.
func (c *ProjectTypeController) detachNames(dto *ProjectActionDto, names []string) types.Output {
	if len(names) == 0 || c.service == nil || c.projectService == nil {
		return nil
	}
	pref := dto.Ctx.Preference
	userID, clientID, workspaceID := ownerFromCtx(dto.Ctx)
	for _, name := range names {
		pt, err := c.service.FindOne(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"name":         name,
		})
		if err != nil || pt == nil {
			return response.BadRequest(
				c.lang.Translate("apps.common.title", nil, pref),
				c.lang.Translate("apps.common.project-type.error.project-not-found", nil, pref),
			)
		}
		_, _ = c.projectService.SoftDelete(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"group_name":   dto.GroupName,
			"service_name": dto.ServiceName,
			"entity_name":  dto.EntityName,
			"entity_id":    dto.EntityId,
			"type_id":      pt.Id,
		})
		if remaining, _ := c.projectService.Count(map[string]any{
			"user_id":      userID,
			"client_id":    clientID,
			"workspace_id": workspaceID,
			"type_id":      pt.Id,
		}); remaining < 1 {
			_, _ = c.service.SoftDelete(map[string]any{"id": pt.Id})
		}
	}
	return nil
}

// diff splits A vs B into (onlyInA, onlyInB, inBoth). Mirrors TS compareProjectTypeNames.
func diff(a, b []string) (onlyInA, onlyInB, inBoth []string) {
	inA := map[string]bool{}
	inB := map[string]bool{}
	for _, v := range a {
		inA[v] = true
	}
	for _, v := range b {
		inB[v] = true
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

// ownerFromCtx extracts (user, client, workspace) from request context.
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
