package app

import (
	"encoding/json"
	"strings"

	"github.com/thescaffold/gox-apps-fuss/app/history"
	"github.com/thescaffold/gox-apps-fuss/app/token"
	"github.com/thescaffold/gox-apps-fuss/app/tokenlog"
)

var unsafeEntities = []string{"fusshistories", "fusstokens", "fusstokenlogs"}

var excludedKeys = []string{
	"id", "status", "created_at", "updated_at", "thumbnail_url",
	"user_id", "client_id", "workspace_id",
	"createdAt", "updatedAt", "thumbnailUrl",
	"userId", "clientId", "workspaceId",
}

type AppService struct {
	tokenEntity    *token.TokenEntity       `inject:""`
	tokenLogEntity *tokenlog.TokenLogEntity `inject:""`
	historyEntity  *history.HistoryEntity   `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

func (s *AppService) HandleCreate(payload map[string]any) error {
	userId, clientId, workspaceId, group, svc, entityName, entityId, meta := extractPayload(payload)
	if !isValidPayload(entityName, entityId, userId) {
		return nil
	}

	metaBytes, _ := json.Marshal(meta)
	tok := &token.Token{
		UserId:      userId,
		ClientId:    clientId,
		WorkspaceId: workspaceId,
		Group:       group,
		Service:     svc,
		EntityId:    entityId,
		EntityName:  entityName,
		Meta:        metaBytes,
	}
	if err := s.tokenEntity.Insert(tok); err != nil {
		return err
	}

	return s.insertTokenLogs(tok.Id, meta)
}

func (s *AppService) HandleUpdate(payload map[string]any) error {
	userId, clientId, workspaceId, group, svc, entityName, entityId, meta := extractPayload(payload)
	if !isValidPayload(entityName, entityId, userId) {
		return nil
	}

	existing, err := s.tokenEntity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "group" = ? AND "service" = ? AND "entity_id" = ? AND "entity_name" = ?`,
		userId, clientId, workspaceId, group, svc, entityId, entityName,
	)
	if err != nil || existing == nil {
		return nil
	}

	// merge meta
	var existingMeta map[string]any
	_ = json.Unmarshal(existing.Meta, &existingMeta)
	for k, v := range meta {
		existingMeta[k] = v
	}
	mergedBytes, _ := json.Marshal(existingMeta)
	existing.Meta = mergedBytes
	if _, err := s.tokenEntity.Update(existing,
		`"id" = ?`, existing.Id,
	); err != nil {
		return err
	}

	// refresh token logs
	if _, err := s.tokenLogEntity.Delete(`"token_id" = ?`, existing.Id); err != nil {
		return err
	}
	return s.insertTokenLogs(existing.Id, meta)
}

func (s *AppService) HandleDelete(payload map[string]any) error {
	userId, clientId, workspaceId, group, svc, entityName, entityId, _ := extractPayload(payload)
	if !isValidPayload(entityName, entityId, userId) {
		return nil
	}

	existing, err := s.tokenEntity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "group" = ? AND "service" = ? AND "entity_id" = ? AND "entity_name" = ?`,
		userId, clientId, workspaceId, group, svc, entityId, entityName,
	)
	if err != nil || existing == nil {
		return nil
	}

	if _, err := s.tokenLogEntity.Delete(`"token_id" = ?`, existing.Id); err != nil {
		return err
	}
	_, err = s.tokenEntity.Delete(`"id" = ?`, existing.Id)
	return err
}

func (s *AppService) Recent(userId, clientId, workspaceId string, page, perPage int) ([]history.History, int64, error) {
	offset := 0
	if page > 1 {
		offset = (page - 1) * perPage
	}
	items, err := s.historyEntity.Find(offset, perPage,
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ?`,
		userId, clientId, workspaceId,
	)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.historyEntity.Count(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ?`,
		userId, clientId, workspaceId,
	)
	return items, total, err
}

func (s *AppService) Search(query, histType, userId, clientId, workspaceId string, page, perPage int) ([]map[string]any, int64, error) {
	// record search history
	h := &history.History{
		UserId:      userId,
		ClientId:    clientId,
		WorkspaceId: workspaceId,
		Type:        histType,
		Query:       query,
	}
	_ = s.historyEntity.Upsert(h)

	offset := 0
	if page > 1 {
		offset = (page - 1) * perPage
	}
	logs, err := s.tokenLogEntity.Find(offset, 0, `"value" LIKE ?`, "%"+query+"%")
	if err != nil {
		return nil, 0, err
	}

	// load associated tokens and build unique entity list
	seen := map[string]struct{}{}
	var results []map[string]any
	for _, l := range logs {
		tok, err := s.tokenEntity.First(`"id" = ?`, l.TokenId)
		if err != nil || tok == nil {
			continue
		}
		var meta map[string]any
		_ = json.Unmarshal(tok.Meta, &meta)
		title, _ := meta["name"].(string)
		if title == "" {
			title, _ = meta["key"].(string)
		}
		if title == "" {
			title, _ = meta["ref"].(string)
		}
		if title == "" {
			continue
		}
		key := tok.EntityName + "-" + tok.EntityId
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, map[string]any{
			"title":      title,
			"desc":       nil,
			"group":      tok.Group,
			"service":    tok.Service,
			"entityId":   tok.EntityId,
			"entityName": tok.EntityName,
		})
	}

	total := int64(len(results))
	start := (page - 1) * perPage
	end := start + perPage
	if start >= int(total) {
		return []map[string]any{}, total, nil
	}
	if end > int(total) {
		end = int(total)
	}
	return results[start:end], total, nil
}

func (s *AppService) Lookup(group, svc, entityName, entityId string) (map[string]any, error) {
	tok, err := s.tokenEntity.First(
		`"group" = ? AND "service" = ? AND "entity_name" = ? AND "entity_id" = ?`,
		group, svc, entityName, entityId,
	)
	if err != nil || tok == nil {
		return nil, nil
	}
	var meta map[string]any
	_ = json.Unmarshal(tok.Meta, &meta)
	return meta, nil
}

// helpers

func (s *AppService) insertTokenLogs(tokenId string, meta map[string]any) error {
	values := getAllValues(meta, excludedKeys)
	for _, v := range values {
		tl := &tokenlog.TokenLog{TokenId: tokenId, Value: v}
		if err := s.tokenLogEntity.Insert(tl); err != nil {
			return err
		}
	}
	return nil
}

func extractPayload(p map[string]any) (userId, clientId, workspaceId, group, service, entityName, entityId string, meta map[string]any) {
	group, _ = p["group"].(string)
	service, _ = p["service"].(string)
	entityName, _ = p["entityName"].(string)
	meta, _ = p["meta"].(map[string]any)
	ctxMap, _ := p["context"].(map[string]any)
	userMap, _ := ctxMap["user"].(map[string]any)
	clientMap, _ := ctxMap["client"].(map[string]any)
	wsMap, _ := ctxMap["workspace"].(map[string]any)
	userId, _ = userMap["id"].(string)
	clientId, _ = clientMap["id"].(string)
	workspaceId, _ = wsMap["id"].(string)
	if meta != nil {
		entityId, _ = meta["id"].(string)
	}
	return
}

func isValidPayload(entityName, entityId, userId string) bool {
	lower := strings.ToLower(entityName)
	for _, unsafe := range unsafeEntities {
		if lower == unsafe {
			return false
		}
	}
	return entityName != "" && entityId != "" && userId != ""
}

func getAllValues(obj map[string]any, omit []string) []string {
	var result []string
	var traverse func(v any)
	traverse = func(v any) {
		switch val := v.(type) {
		case map[string]any:
			for k, child := range val {
				if !contains(omit, k) {
					traverse(child)
				}
			}
		case string:
			if len(val) > 0 && len(val) < 255 {
				result = append(result, val)
			}
		}
	}
	traverse(obj)
	return result
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
