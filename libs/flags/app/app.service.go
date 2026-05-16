package app

import (
	environmentpkg "github.com/thescaffold/gox-apps/libs/flags/app/environment"
	environmenttypepkg "github.com/thescaffold/gox-apps/libs/flags/app/environmenttype"
	flagpkg "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	flaglogpkg "github.com/thescaffold/gox-apps/libs/flags/app/flaglog"
)

// AppService implements register/log/status/limit, mirroring
// ntx-apps/libs/flags/src/app.service.ts. It persists flag definitions and
// per-flag log rows; status/limit aggregate over flag.limit + sum(log.limit).
type AppService struct {
	flagEntity        *flagpkg.FlagEntity                       `inject:""`
	flagLogEntity     *flaglogpkg.FlagLogEntity                 `inject:""`
	environmentEntity *environmentpkg.EnvironmentEntity         `inject:""`
	envTypeEntity     *environmenttypepkg.EnvironmentTypeEntity `inject:""`
}

// GetHello mirrors TS AppService.getHello().
func (s *AppService) GetHello() string { return "Hello World!" }

// FlagOwner is the user/client/workspace tuple a flag is registered against.
type FlagOwner struct {
	UserId      string
	ClientId    string
	WorkspaceId string
}

// Register upserts each flag definition for the (env, environmentType, owner)
// tuple. Mirrors TS appService.register(). The environment id is resolved
// via Environment find-or-create on (name, envTypeId) — repeated registers
// for the same name share a stable id so downstream Log/Status/Limit lookups
// resolve consistently regardless of which caller seeded the env.
func (s *AppService) Register(envType, env EnvironmentRef, owner FlagOwner, flags []FlagDef) ([]flagpkg.Flag, error) {
	envID := s.envIDForResolve(envType, env, owner)
	out := make([]flagpkg.Flag, 0, len(flags))
	for _, f := range flags {
		existing, _ := s.flagEntity.First(
			`environment_id = ? AND user_id = ? AND client_id = ? AND workspace_id = ? AND name = ?`,
			envID, owner.UserId, owner.ClientId, owner.WorkspaceId, f.Name,
		)
		if existing != nil {
			existing.Limit = f.Limit
			existing.Priority = f.Priority
			existing.Level = orDefault(f.Level, FlagLevelUser)
			existing.Meta = f.Meta
			existing.Status = f.Status
			if _, err := s.flagEntity.Update(existing, `id = ?`, existing.Id); err != nil {
				return nil, err
			}
			out = append(out, *existing)
			continue
		}
		row := &flagpkg.Flag{
			UserId:        owner.UserId,
			ClientId:      owner.ClientId,
			WorkspaceId:   owner.WorkspaceId,
			EnvironmentId: envID,
			Name:          f.Name,
			Limit:         f.Limit,
			Priority:      f.Priority,
			Level:         orDefault(f.Level, FlagLevelUser),
			Meta:          f.Meta,
			Status:        f.Status,
		}
		if err := s.flagEntity.Insert(row); err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	return out, nil
}

// LogQuery is the params passed to Log.
type LogQuery struct {
	UserId      string
	ClientId    string
	WorkspaceId string
	Name        string
	Level       string
}

// Log records a single flag-log row by incrementing usage on the matching
// (env, owner, name, level) flag. limit defaults to 1.
// Mirrors TS appService.log(): writes one FlagLog row with the supplied limit.
func (s *AppService) Log(env EnvironmentRef, q LogQuery, limit int) (*flaglogpkg.FlagLog, error) {
	flag, err := s.flag(env, q)
	if err != nil || flag == nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 1
	}
	row := &flaglogpkg.FlagLog{FlagId: flag.Id, Limit: limit}
	if err := s.flagLogEntity.Insert(row); err != nil {
		return nil, err
	}
	return row, nil
}

// Status returns true when the flag exists, is Active, and its summed log
// usage has not yet reached the flag's limit. Mirrors the *intent* of TS
// appService.status(); the literal TS implementation has two bugs — a missing
// `return` (status always undefined) and an inverted comparison (`limit <= usage`).
// gox implements the intent because the broken TS form is unusable.
func (s *AppService) Status(env EnvironmentRef, q LogQuery) (bool, error) {
	flag, err := s.flag(env, q)
	if err != nil || flag == nil {
		return false, err
	}
	if flag.Status != nil && *flag.Status != "active" {
		return false, nil
	}
	if flag.Limit <= 0 {
		// 0 means unlimited.
		return true, nil
	}
	used, err := s.sumFlagLogLimits(flag.Id)
	if err != nil {
		return false, err
	}
	return used < flag.Limit, nil
}

// LimitResult is the (allowed, limit, usage) triple returned by Limit. JSON
// tags match the TS response keys.
type LimitResult struct {
	Allowed bool `json:"allowed"`
	Limit   int  `json:"limit"`
	Usage   int  `json:"usage"`
}

// Limit returns the allowed/limit/usage triple. Mirrors the intent of TS
// appService.limit(): allowed is true when the flag is Active AND usage is
// below limit; usage is the sum of FlagLog.limit values for the flag.
func (s *AppService) Limit(env EnvironmentRef, q LogQuery) (*LimitResult, error) {
	flag, err := s.flag(env, q)
	if err != nil {
		return nil, err
	}
	if flag == nil {
		return &LimitResult{Allowed: false, Limit: 0, Usage: 0}, nil
	}
	active := flag.Status == nil || *flag.Status == "active"
	if !active {
		return &LimitResult{Allowed: false, Limit: flag.Limit, Usage: 0}, nil
	}
	used, err := s.sumFlagLogLimits(flag.Id)
	if err != nil {
		return nil, err
	}
	allowed := flag.Limit <= 0 || used < flag.Limit
	return &LimitResult{Allowed: allowed, Limit: flag.Limit, Usage: used}, nil
}

// sumFlagLogLimits computes SUM(limit) for the FlagLog rows of a given flag.
// Mirrors TS flagLogService.sum('limit', { flagId }). goose's Entity doesn't
// expose SUM directly, so we Find + sum in-memory; usage counts are small
// per-flag so this is acceptable.
func (s *AppService) sumFlagLogLimits(flagID string) (int, error) {
	rows, err := s.flagLogEntity.Find(0, 0, `flag_id = ?`, flagID)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, r := range rows {
		total += r.Limit
	}
	return total, nil
}

// flag resolves the matching Flag row for the (env, owner, name, level) tuple.
// Read-path: when the env ref carries no explicit ID, prefer the resolved
// environment row id (find-by-name); fall back to the name only as a last
// resort so Log/Status/Limit still work against legacy rows seeded before
// the EnvironmentService was wired.
func (s *AppService) flag(env EnvironmentRef, q LogQuery) (*flagpkg.Flag, error) {
	level := orDefault(q.Level, FlagLevelUser)
	envID := envIDFor(env)
	if (env.ID == nil || *env.ID == "") && s.environmentEntity != nil && env.Name != "" {
		row, _ := s.environmentEntity.First(`"name" = ?`, env.Name)
		if row != nil {
			envID = row.Id
		}
	}
	return s.flagEntity.First(
		`environment_id = ? AND user_id = ? AND client_id = ? AND workspace_id = ? AND name = ? AND level = ?`,
		envID, q.UserId, q.ClientId, q.WorkspaceId, q.Name, level,
	)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// envIDFor resolves an EnvironmentRef to a string id. When the ref carries an
// explicit ID, use it; otherwise fall back to the name as a deterministic
// pseudo-id. Prefer envIDForResolve when an AppService is available so the
// lookup goes through the real Environment entity.
func envIDFor(env EnvironmentRef) string {
	if env.ID != nil && *env.ID != "" {
		return *env.ID
	}
	return env.Name
}

// envIDForResolve mirrors TS environmentService.findOrCreate({name}) — when
// the EnvironmentRef carries an explicit ID, use it; otherwise look up by
// name and create a row if missing so subsequent lookups return the same id.
// The (envType, owner) tuple is required to attribute the row; passes
// through to envIDFor when the AppService doesn't have an Environment entity
// wired (tests, host bootstrap).
func (s *AppService) envIDForResolve(envType, env EnvironmentRef, owner FlagOwner) string {
	if env.ID != nil && *env.ID != "" {
		return *env.ID
	}
	if s.environmentEntity == nil || env.Name == "" {
		return env.Name
	}
	// Resolve the environment-type id (find-or-create by name).
	typeID := ""
	if s.envTypeEntity != nil && envType.Name != "" {
		etRow, _ := s.envTypeEntity.First(`"name" = ?`, envType.Name)
		if etRow == nil {
			et := &environmenttypepkg.EnvironmentType{Name: envType.Name, Category: "flag"}
			_ = s.envTypeEntity.Insert(et)
			etRow = et
		}
		typeID = etRow.Id
	}
	row, _ := s.environmentEntity.First(
		`"name" = ? AND "type_id" = ?`, env.Name, typeID,
	)
	if row != nil {
		return row.Id
	}
	created := &environmentpkg.Environment{
		Name:        env.Name,
		TypeId:      typeID,
		UserId:      owner.UserId,
		ClientId:    owner.ClientId,
		WorkspaceId: owner.WorkspaceId,
	}
	if err := s.environmentEntity.Insert(created); err != nil {
		// Fall back to deterministic name-as-id so callers keep working.
		return env.Name
	}
	return created.Id
}
