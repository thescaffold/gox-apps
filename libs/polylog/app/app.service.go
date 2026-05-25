package app

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/thescaffold/gox-apps/libs/polylog/app/channel"
	polylogconfig "github.com/thescaffold/gox-apps/libs/polylog/app/config"
	"github.com/thescaffold/gox-apps/libs/polylog/app/event"
	"github.com/thescaffold/gox-apps/libs/polylog/app/eventlog"
	"github.com/thescaffold/gox-apps/libs/polylog/app/sink"
	"github.com/thescaffold/gox-apps/libs/polylog/app/sinktype"
	"github.com/thescaffold/gox-apps/libs/polylog/app/source"
	"github.com/thescaffold/gox-apps/libs/polylog/app/sourcetype"
	"github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// asterisk mirrors TS ASTERISK = '965c068d-52d9-4cf4-907f-8e453f20074a' from
// ntx-apps/libs/polylog/src/app.config.ts. Used as a wildcard source/channel id.
const asterisk = "965c068d-52d9-4cf4-907f-8e453f20074a"

// AppService surfaces polylog ingest/config operations.
// Mirrors ntx-apps/libs/polylog/src/app.service.ts.
type AppService struct {
	pipelineService  *pipeline.PipelineService    `inject:""`
	configService    *polylogconfig.ConfigService `inject:""`
	sourceEntity     *source.SourceEntity         `inject:""`
	sourceTypeEntity *sourcetype.SourceTypeEntity `inject:""`
	sinkEntity       *sink.SinkEntity             `inject:""`
	sinkTypeEntity   *sinktype.SinkTypeEntity     `inject:""`
	channelEntity    *channel.ChannelEntity       `inject:""`
	eventEntity      *event.EventEntity           `inject:""`
	eventLogEntity   *eventlog.EventLogEntity     `inject:""`
}

// Housekeep deletes Event/EventLog rows older than cutoff. Mirrors TS
// hourly heartbeat housekeeping.
func (s *AppService) Housekeep(cutoff time.Time) {
	if s.eventEntity != nil {
		_, _ = s.eventEntity.Delete(`"created_at" < ?`, cutoff)
	}
	if s.eventLogEntity != nil {
		_, _ = s.eventLogEntity.Delete(`"created_at" < ?`, cutoff)
	}
}

func (s *AppService) GetHello() string { return "Hello World!" }

// ProcessEvent hands a stored Event off to PipelineService.Process. Used by
// the `queue/apps/polylog/ingest` queue consumer registered in index.go.
func (s *AppService) ProcessEvent(ev *event.Event) error {
	if s.pipelineService == nil || ev == nil {
		return nil
	}
	return s.pipelineService.Process(ev)
}

// IngestBatch processes a batch of items.
// Mirrors TS AppService.ingestBatch().
func (s *AppService) IngestBatch(items []pipeline.Item, ctx ntxctx.NTXContext) ([]any, error) {
	out := make([]any, 0, len(items))
	for _, it := range items {
		ev, err := s.pipelineService.Ingest(it, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, nil
}

// Ingest processes a single item. Mirrors TS AppService.ingest().
func (s *AppService) Ingest(item pipeline.Item, ctx ntxctx.NTXContext) (any, error) {
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// IngestSourceAlt is `POST /ingest/:id` — the :id is the source id; body+queries are merged.
func (s *AppService) IngestSourceAlt(body json.RawMessage, sourceID string, properties map[string]string, ctx ntxctx.NTXContext) (any, error) {
	item := pipeline.Item{Source: sourceID, Payload: body}
	if t, ok := properties["type"]; ok {
		item.Type = t
	}
	if c, ok := properties["category"]; ok {
		item.Category = c
	}
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// IngestSource is `POST /ingest/source/:id`. TS body sets entityName="source".
func (s *AppService) IngestSource(body json.RawMessage, sourceID string, properties map[string]string, ctx ntxctx.NTXContext) (any, error) {
	item := pipeline.Item{Source: sourceID, EntityName: "source", Payload: body}
	if t, ok := properties["type"]; ok {
		item.Type = t
	}
	if c, ok := properties["category"]; ok {
		item.Category = c
	}
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// IngestChannel is `POST /ingest/channel/:id`. TS body sets entityName="channel".
func (s *AppService) IngestChannel(body json.RawMessage, channelID string, properties map[string]string, ctx ntxctx.NTXContext) (any, error) {
	item := pipeline.Item{Source: channelID, EntityName: "channel", Payload: body}
	if t, ok := properties["type"]; ok {
		item.Type = t
	}
	if c, ok := properties["category"]; ok {
		item.Category = c
	}
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// SetConfig mirrors TS AppService.setConfig — a 3-way upsert across
// Source / Sink / Channel. The source is matched by (typeId, name); the sink
// (when provided) by (typeId, name); the channel by
// (userId, clientId, workspaceId, sourceId, sinkId). When sink.id or source.id
// is supplied directly, it's updated in-place. Returns the Channel row.
func (s *AppService) SetConfig(dto *SetConfigDto) (any, error) {
	if dto == nil {
		return nil, nil
	}

	srcType, _ := s.sourceTypeEntity.First(`"key" = ?`, dto.SourceType.Key)
	if srcType == nil {
		return nil, nil
	}

	// Source upsert: id-driven update, then (typeId,name) find, else insert.
	newSource, err := s.upsertSource(dto, srcType.Id)
	if err != nil || newSource == nil {
		return nil, err
	}

	// Sink upsert (optional).
	var newSink *sink.Sink
	if dto.Sink != nil && dto.SinkType != nil {
		skType, _ := s.sinkTypeEntity.First(`"key" = ?`, dto.SinkType.Key)
		if skType != nil {
			newSink, err = s.upsertSink(dto, skType.Id)
			if err != nil {
				return nil, err
			}
		}
	}

	// Channel upsert: match (userId,clientId,workspaceId,sourceId,sinkId).
	var sinkId *string
	if newSink != nil {
		sid := newSink.Id
		sinkId = &sid
	}
	var existing *channel.Channel
	if sinkId != nil {
		existing, _ = s.channelEntity.First(
			`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "sink_id" = ?`,
			dto.UserId, dto.ClientId, dto.WorkspaceId, newSource.Id, *sinkId,
		)
	} else {
		existing, _ = s.channelEntity.First(
			`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "sink_id" IS NULL`,
			dto.UserId, dto.ClientId, dto.WorkspaceId, newSource.Id,
		)
	}

	if existing != nil {
		if sinkId != nil && (existing.SinkId == nil || *existing.SinkId != *sinkId) {
			existing.SinkId = sinkId
			_, _ = s.channelEntity.Update(existing, `"id" = ?`, existing.Id)
		}
		return existing, nil
	}

	cat := dto.Category
	if cat == "" {
		cat = newSource.Category
	}
	ch := &channel.Channel{
		UserId:      dto.UserId,
		ClientId:    dto.ClientId,
		WorkspaceId: dto.WorkspaceId,
		Category:    cat,
		SourceId:    newSource.Id,
		SinkId:      sinkId,
		Type:        dto.Type,
	}
	if err := s.channelEntity.Insert(ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *AppService) upsertSource(dto *SetConfigDto, typeId string) (*source.Source, error) {
	src := dto.Source
	if src.Id != nil && *src.Id != "" {
		existing, _ := s.sourceEntity.First(`"id" = ?`, *src.Id)
		if existing == nil {
			return nil, nil
		}
		existing.TypeId = typeId
		existing.Category = src.Category
		existing.Name = src.Name
		existing.Desc = src.Desc
		existing.Meta = src.Meta
		existing.Status = src.Status
		_, _ = s.sourceEntity.Update(existing, `"id" = ?`, existing.Id)
		return existing, nil
	}
	existing, _ := s.sourceEntity.First(`"type_id" = ? AND "name" = ?`, typeId, src.Name)
	if existing != nil {
		return existing, nil
	}
	row := &source.Source{
		UserId:      dto.UserId,
		ClientId:    dto.ClientId,
		WorkspaceId: dto.WorkspaceId,
		Category:    src.Category,
		Key:         src.Key,
		Visibility:  src.Visibility,
		TypeId:      typeId,
		Name:        src.Name,
		Desc:        src.Desc,
		Meta:        src.Meta,
		Status:      src.Status,
	}
	if err := s.sourceEntity.Insert(row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *AppService) upsertSink(dto *SetConfigDto, typeId string) (*sink.Sink, error) {
	sk := dto.Sink
	if sk.Id != nil && *sk.Id != "" {
		existing, _ := s.sinkEntity.First(`"id" = ?`, *sk.Id)
		if existing == nil {
			return nil, nil
		}
		existing.TypeId = typeId
		existing.Category = sk.Category
		existing.Name = sk.Name
		existing.Desc = sk.Desc
		existing.Meta = sk.Meta
		_, _ = s.sinkEntity.Update(existing, `"id" = ?`, existing.Id)
		return existing, nil
	}
	existing, _ := s.sinkEntity.First(`"type_id" = ? AND "name" = ?`, typeId, sk.Name)
	if existing != nil {
		return existing, nil
	}
	row := &sink.Sink{
		UserId:      dto.UserId,
		ClientId:    dto.ClientId,
		WorkspaceId: dto.WorkspaceId,
		Category:    sk.Category,
		Key:         sk.Key,
		Visibility:  sk.Visibility,
		TypeId:      typeId,
		Name:        sk.Name,
		Desc:        sk.Desc,
		Meta:        sk.Meta,
	}
	if err := s.sinkEntity.Insert(row); err != nil {
		return nil, err
	}
	return row, nil
}

// dottedTypeVariations expands a dotted type string into the 2^n wildcard
// combinations used by GetConfig channel matching, mirroring TS
// generateDottedStringVariations: each segment is either itself or the ASTERISK
// sentinel. e.g. "a.b" → ["a.b", "<*>.b", "a.<*>", "<*>.<*>"]. (The previous
// cumulative-prefix version did not match TS and missed wildcard channels.)
func dottedTypeVariations(t string) []string {
	segments := strings.Split(t, ".")
	total := 1 << len(segments)
	out := make([]string, 0, total)
	for i := 0; i < total; i++ {
		combo := make([]string, len(segments))
		for j, seg := range segments {
			if i&(1<<j) != 0 {
				combo[j] = asterisk
			} else {
				combo[j] = seg
			}
		}
		out = append(out, strings.Join(combo, "."))
	}
	return out
}

// GetConfig mirrors TS AppService.getConfig — channel lookup driven by
// (entityId, entityName, type, category). For source entities, builds OR'd
// filters across the dotted-type variations + ASTERISK fallbacks. For other
// entity names, simply fetches the channel by id.
func (s *AppService) GetConfig(userID, clientID, workspaceID, entityID, entityName, typ, category string, ctx ntxctx.NTXContext) (any, error) {
	if userID == "" {
		userID = ctx.UserID
	}
	if clientID == "" {
		clientID = ctx.ClientID
	}
	if workspaceID == "" {
		workspaceID = ctx.WorkspaceID
	}

	if entityName != "source" {
		row, _ := s.channelEntity.First(`"id" = ?`, entityID)
		if row == nil {
			return []*channel.Channel{}, nil
		}
		return []*channel.Channel{row}, nil
	}

	variations := dottedTypeVariations(typ)
	var orParts []string
	var args []any

	add := func(sourceId, sinkType string) {
		orParts = append(orParts,
			`("user_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "category" = ? AND "type" = ?)`)
		args = append(args, userID, workspaceID, sourceId, category, sinkType)
	}
	// ASTERISK channel-type matches.
	add(entityID, asterisk)
	add(asterisk, asterisk)
	// dotted variations across (sourceId == entityID || sourceId == ASTERISK).
	for _, v := range variations {
		add(entityID, v)
		add(asterisk, v)
	}

	if len(orParts) == 0 {
		return []*channel.Channel{}, nil
	}
	where := strings.Join(orParts, " OR ")
	rows, err := s.channelEntity.Find(0, 0, where, args...)
	if err != nil {
		return nil, err
	}
	// Deduplicate by id, mirroring TS .filter((v,i,a)=>a.findIndex(t=>t.id===v.id)===i).
	seen := map[string]bool{}
	out := make([]channel.Channel, 0, len(rows))
	for _, r := range rows {
		if seen[r.Id] {
			continue
		}
		seen[r.Id] = true
		out = append(out, r)
	}
	return out, nil
}

// UpdateConfig mirrors TS AppService.updateConfig — patches a Source row's
// status by id. Returns nil when the source does not exist, so the controller
// can emit the TS-equivalent invalid-request envelope.
func (s *AppService) UpdateConfig(id string, status string, ctx ntxctx.NTXContext) (any, error) {
	if id == "" {
		return nil, nil
	}
	existing, _ := s.sourceEntity.First(`"id" = ?`, id)
	if existing == nil {
		return nil, nil
	}
	st := status
	existing.Status = &st
	if _, err := s.sourceEntity.Update(existing, `"id" = ?`, existing.Id); err != nil {
		return nil, err
	}
	return map[string]any{"id": id}, nil
}

// LegacyConfigUpsert remains for downstream callers that still drive the
// PolylogConfigs table directly (rate-limited usage from telemetry replays).
// Not exposed via HTTP; safe to deprecate once those replays migrate.
func (s *AppService) LegacyConfigUpsert(body map[string]any, ctx ntxctx.NTXContext) (any, error) {
	entityId, entityName, typ, category := configKeyFromBody(body)
	value, _ := json.Marshal(body["value"])
	var status *string
	if v, ok := body["status"].(string); ok {
		status = &v
	}
	return s.configService.Upsert(
		ctx.UserID, ctx.ClientID, ctx.WorkspaceID,
		entityId, entityName, typ, category,
		value, status,
	)
}

func configKeyFromBody(body map[string]any) (entityId, entityName, typ, category string) {
	if v, ok := body["entityId"].(string); ok {
		entityId = v
	}
	if v, ok := body["entityName"].(string); ok {
		entityName = v
	}
	if v, ok := body["type"].(string); ok {
		typ = v
	}
	if v, ok := body["category"].(string); ok {
		category = v
	}
	return
}

// CreateRootSource mirrors TS AppService.createRootSource(): find the
// SourceType by typeKey, then findOrCreate a Source for the root user/client/
// workspace (env vars ROOT_USER_ID/ROOT_CLIENT_ID/ROOT_WORKSPACE_ID). When an
// id is supplied the lookup is by id; otherwise by (rootUser/client/workspace,
// typeId, category, key) tuple.
func (s *AppService) CreateRootSource(dto *UpsertRootSourceDto) (any, error) {
	if dto == nil {
		return nil, nil
	}
	srcType, _ := s.sourceTypeEntity.First(`"key" = ?`, dto.TypeKey)
	if srcType == nil {
		return nil, nil
	}
	rootUser := os.Getenv("ROOT_USER_ID")
	rootClient := os.Getenv("ROOT_CLIENT_ID")
	rootWorkspace := os.Getenv("ROOT_WORKSPACE_ID")

	var existing *source.Source
	if dto.Id != nil && *dto.Id != "" {
		existing, _ = s.sourceEntity.First(`"id" = ?`, *dto.Id)
	} else {
		key := ""
		if dto.Key != nil {
			key = *dto.Key
		}
		existing, _ = s.sourceEntity.First(
			`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "type_id" = ? AND "category" = ? AND "key" = ?`,
			rootUser, rootClient, rootWorkspace, srcType.Id, dto.Category, key,
		)
	}
	if existing != nil {
		return existing, nil
	}
	row := &source.Source{
		UserId:      rootUser,
		ClientId:    rootClient,
		WorkspaceId: rootWorkspace,
		Category:    dto.Category,
		Key:         dto.Key,
		Visibility:  dto.Visibility,
		TypeId:      srcType.Id,
		Name:        dto.Name,
		Desc:        dto.Desc,
		Meta:        dto.Meta,
		Status:      dto.Status,
	}
	if err := s.sourceEntity.Insert(row); err != nil {
		return nil, err
	}
	return row, nil
}

// CreateRootSink mirrors TS AppService.createRootSink(): find the SinkType by
// sinkType.key, updateOrCreate the Sink for the root user/client/workspace,
// then updateOrCreate a Channel linking source=ASTERISK to the new sink.
func (s *AppService) CreateRootSink(dto *UpsertRootSinkDto) (any, error) {
	if dto == nil {
		return nil, nil
	}
	skType, _ := s.sinkTypeEntity.First(`"key" = ?`, dto.SinkType.Key)
	if skType == nil {
		return nil, nil
	}
	rootUser := os.Getenv("ROOT_USER_ID")
	rootClient := os.Getenv("ROOT_CLIENT_ID")
	rootWorkspace := os.Getenv("ROOT_WORKSPACE_ID")

	var existingSink *sink.Sink
	if dto.Sink.Id != nil && *dto.Sink.Id != "" {
		existingSink, _ = s.sinkEntity.First(`"id" = ?`, *dto.Sink.Id)
	} else {
		key := ""
		if dto.Sink.Key != nil {
			key = *dto.Sink.Key
		}
		existingSink, _ = s.sinkEntity.First(
			`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "type_id" = ? AND "category" = ? AND "key" = ?`,
			rootUser, rootClient, rootWorkspace, skType.Id, dto.Sink.Category, key,
		)
	}
	if existingSink != nil {
		existingSink.Name = dto.Sink.Name
		existingSink.Desc = dto.Sink.Desc
		existingSink.Meta = dto.Sink.Meta
		_, _ = s.sinkEntity.Update(existingSink, `"id" = ?`, existingSink.Id)
	} else {
		existingSink = &sink.Sink{
			UserId:      rootUser,
			ClientId:    rootClient,
			WorkspaceId: rootWorkspace,
			Category:    dto.Sink.Category,
			Key:         dto.Sink.Key,
			Visibility:  dto.Sink.Visibility,
			TypeId:      skType.Id,
			Name:        dto.Sink.Name,
			Desc:        dto.Sink.Desc,
			Meta:        dto.Sink.Meta,
		}
		if err := s.sinkEntity.Insert(existingSink); err != nil {
			return nil, err
		}
	}

	// link a channel: sourceId=ASTERISK, sinkId=sink.id, type=dto.Type.
	existingChannel, _ := s.channelEntity.First(
		`"user_id" = ? AND "client_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "sink_id" = ? AND "type" = ?`,
		rootUser, rootClient, rootWorkspace, asterisk, existingSink.Id, dto.Type,
	)
	if existingChannel != nil {
		return existingChannel, nil
	}
	typeCopy := dto.Type
	ch := &channel.Channel{
		UserId:      rootUser,
		ClientId:    rootClient,
		WorkspaceId: rootWorkspace,
		Category:    dto.Category,
		SourceId:    asterisk,
		SinkId:      &existingSink.Id,
		Type:        &typeCopy,
	}
	if err := s.channelEntity.Insert(ch); err != nil {
		return nil, err
	}
	return ch, nil
}
