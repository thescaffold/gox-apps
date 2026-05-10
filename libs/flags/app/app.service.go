package app

import (
	flagpkg "github.com/thescaffold/gox-apps-flags/app/flag"
	flaglogpkg "github.com/thescaffold/gox-apps-flags/app/flaglog"
)

// AppService implements register/log/status/limit, mirroring
// ntx-apps/libs/flags/src/app.service.ts. It persists flag definitions and
// per-flag log rows; status/limit aggregate over flag.limit + flaglog count.
type AppService struct {
	flagEntity    *flagpkg.FlagEntity       `inject:""`
	flagLogEntity *flaglogpkg.FlagLogEntity `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

// FlagOwner is the user/client/workspace tuple a flag is registered against.
type FlagOwner struct {
	UserId      string
	ClientId    string
	WorkspaceId string
}

// Register upserts each flag definition for the (env, environmentType, owner)
// tuple. Mirrors TS appService.register(). environmentId is resolved by name
// (TODO: requires Environment service; falls back to env.Name as id for now).
func (s *AppService) Register(envType, env EnvironmentRef, owner FlagOwner, flags []FlagDef) ([]flagpkg.Flag, error) {
	envID := envIDFor(env)
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
// Mirrors TS appService.log().
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

// Status returns true when the flag exists and is currently below its limit.
// Mirrors TS appService.status() per-name evaluation.
func (s *AppService) Status(env EnvironmentRef, q LogQuery) (bool, error) {
	flag, err := s.flag(env, q)
	if err != nil || flag == nil {
		return false, err
	}
	if flag.Status != nil && *flag.Status == "off" {
		return false, nil
	}
	if flag.Limit <= 0 {
		// 0 means unlimited
		return true, nil
	}
	used, err := s.flagLogEntity.Count(`flag_id = ?`, flag.Id)
	if err != nil {
		return false, err
	}
	return used < int64(flag.Limit), nil
}

// LimitResult is the (allowed, limit, usage) triple returned by Limit.
type LimitResult struct {
	Allowed bool `json:"allowed"`
	Limit   int  `json:"limit"`
	Usage   int  `json:"usage"`
}

// Limit returns the allowed/limit/usage triple. Mirrors TS appService.limit().
func (s *AppService) Limit(env EnvironmentRef, q LogQuery) (*LimitResult, error) {
	flag, err := s.flag(env, q)
	if err != nil {
		return nil, err
	}
	if flag == nil {
		return &LimitResult{Allowed: false, Limit: 0, Usage: 0}, nil
	}
	used, err := s.flagLogEntity.Count(`flag_id = ?`, flag.Id)
	if err != nil {
		return nil, err
	}
	allowed := flag.Limit <= 0 || used < int64(flag.Limit)
	return &LimitResult{Allowed: allowed, Limit: flag.Limit, Usage: int(used)}, nil
}

// flag resolves the matching Flag row for the (env, owner, name, level) tuple.
func (s *AppService) flag(env EnvironmentRef, q LogQuery) (*flagpkg.Flag, error) {
	level := orDefault(q.Level, FlagLevelUser)
	envID := envIDFor(env)
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
// pseudo-id (TODO: replace with EnvironmentService lookup).
func envIDFor(env EnvironmentRef) string {
	if env.ID != nil && *env.ID != "" {
		return *env.ID
	}
	return env.Name
}
