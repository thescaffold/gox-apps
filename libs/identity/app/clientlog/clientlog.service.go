// Package clientlog hosts the ClientLog persistence layer + the OAuth-flow
// helpers used by identity's pkg/oauth.controller.go (server/device/web
// initiate + auth-code redeem).
package clientlog

import (
	"encoding/json"
	"time"

	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// ClientLogService exposes high-level operations used by the OAuth controllers.
// Mirrors the TS clientLogService surface where every step of the OAuth flow
// persists a row (initiate → auth-code → access-token), letting the host
// audit / replay the dance.
type ClientLogService struct {
	entity *ClientLogEntity `inject:""`
}

// Entity returns the underlying entity for callers needing low-level access.
func (s *ClientLogService) Entity() *ClientLogEntity { return s.entity }

// CreateInitiate persists an "initiate" ClientLog row for the supplied client
// and returns the inserted row. The row's id doubles as the OAuth sessionId
// downstream callers reference. Mirrors TS OauthServerInitiate handler.
func (s *ClientLogService) CreateInitiate(clientId, state, scope, kind string) (*ClientLog, error) {
	if s.entity == nil {
		return nil, nil
	}
	status := "initiate"
	reqJSON, _ := json.Marshal(map[string]any{
		"state": state, "scope": scope, "kind": kind,
	})
	row := &ClientLog{
		ClientId: clientId,
		Event:    "oauth:" + kind + ":initiate",
		Request:  reqJSON,
		Status:   &status,
	}
	if err := s.entity.Insert(row); err != nil {
		return nil, err
	}
	return row, nil
}

// SetAuthCode mints a fresh auth-code for the supplied sessionId, stores it
// in the row's response payload, and returns the code. Mirrors TS
// OauthServerAuthCode handler.
func (s *ClientLogService) SetAuthCode(sessionId, userId, workspaceId string) (string, error) {
	if s.entity == nil {
		return "", nil
	}
	row, _ := s.entity.First(`"id" = ?`, sessionId)
	if row == nil {
		return "", nil
	}
	authCode := utils.Reference("AUC", 28)
	now := time.Now().UTC()
	resp := map[string]any{
		"authCode":    authCode,
		"userId":      userId,
		"workspaceId": workspaceId,
		"issuedAt":    now,
	}
	respJSON, _ := json.Marshal(resp)
	status := "auth-code"
	row.Response = respJSON
	row.Status = &status
	if _, err := s.entity.Update(row, `"id" = ?`, row.Id); err != nil {
		return "", err
	}
	return authCode, nil
}

// RedeemAuthCode swaps a (sessionId, authCode) pair for the recorded
// (userId, workspaceId, clientId) triple. Returns nil when the code doesn't
// match or has already been redeemed. Single-use — sets status="access-token"
// on success so subsequent redeems fail.
func (s *ClientLogService) RedeemAuthCode(sessionId, authCode string) (*RedeemedClientLog, error) {
	if s.entity == nil {
		return nil, nil
	}
	row, _ := s.entity.First(`"id" = ?`, sessionId)
	if row == nil {
		return nil, nil
	}
	if row.Status != nil && *row.Status == "access-token" {
		return nil, nil
	}
	var resp map[string]any
	if len(row.Response) > 0 {
		_ = json.Unmarshal(row.Response, &resp)
	}
	stored, _ := resp["authCode"].(string)
	if stored == "" || stored != authCode {
		return nil, nil
	}
	userId, _ := resp["userId"].(string)
	workspaceId, _ := resp["workspaceId"].(string)
	consumed := "access-token"
	row.Status = &consumed
	_, _ = s.entity.Update(row, `"id" = ?`, row.Id)
	return &RedeemedClientLog{
		SessionId:   row.Id,
		ClientId:    row.ClientId,
		UserId:      userId,
		WorkspaceId: workspaceId,
	}, nil
}

// RedeemedClientLog is the return shape of RedeemAuthCode.
type RedeemedClientLog struct {
	SessionId   string
	ClientId    string
	UserId      string
	WorkspaceId string
}
