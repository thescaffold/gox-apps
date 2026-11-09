package user

import "github.com/thescaffold/gox-packages/libs/core/events"

// EventUserCreated is published once a new account exists: after sign-up by
// password, a passwordless code, a social provider or SSO, and after
// AuthService.RegisterUser. Hosts subscribe to it for work that must happen
// eagerly for every new person (a mailbox, a wallet, a welcome message)
// instead of on their first request.
const EventUserCreated = "apps.identity.user.created"

// Created publishes EventUserCreated for u on tracker (the process's bus):
// {userId, clientId, workspaceId, ref, type, name, email, method}. workspaceId
// is the workspace the sign-up bootstrapped, "" when there was none; method
// names the path ("password", "passwordless", "provider:<name>", "sso",
// "register"). A nil tracker or user publishes nothing.
func Created(tracker *events.TrackerService, u *User, clientID, workspaceID, method string) {
	if tracker == nil || u == nil || u.Id == "" {
		return
	}
	p := map[string]any{
		"userId":      u.Id,
		"clientId":    clientID,
		"workspaceId": workspaceID,
		"name":        u.Name,
		"email":       u.Email,
		"method":      method,
	}
	if u.Ref != nil {
		p["ref"] = *u.Ref
	}
	if u.Type != nil {
		p["type"] = *u.Type
	}
	tracker.Message(EventUserCreated, p)
}
