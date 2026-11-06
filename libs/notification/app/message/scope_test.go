package message

import (
	"testing"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
)

func create(t *testing.T, cfg crud.Config[Message, CreateMessageDto, UpdateMessageDto], dto *CreateMessageDto, ctx ntxctx.NTXContext) (*CreateMessageDto, error) {
	t.Helper()
	out, err := cfg.Morphs[crud.BeforeCreate](dto, ctx)
	if err != nil {
		return nil, err
	}
	return out.(*CreateMessageDto), nil
}

func TestUnscopedByDefault(t *testing.T) {
	ownerScoped.Store(false)
	cfg := config()
	if cfg.WorkspaceScoped || cfg.OwnerScoped {
		t.Fatal("the message API is scoped without ScopeToOwner")
	}
	// Ids from verified claims fill the owner when the caller maps are absent.
	got, err := create(t, cfg, &CreateMessageDto{}, ntxctx.NTXContext{UserID: "u1", WorkspaceID: "w1", ClientID: "c1"})
	if err != nil || got.UserId == nil || *got.UserId != "u1" || *got.WorkspaceId != "w1" || *got.ClientId != "c1" {
		t.Fatalf("owner from claims: %+v %v", got, err)
	}
	// Without a caller, as before, a message may name its recipient.
	other := "u2"
	got, err = create(t, cfg, &CreateMessageDto{UserId: &other}, ntxctx.NTXContext{})
	if err != nil || *got.UserId != "u2" {
		t.Fatalf("unscoped create: %+v %v", got, err)
	}
}

func TestScopeToOwner(t *testing.T) {
	ScopeToOwner()
	t.Cleanup(func() { ownerScoped.Store(false) })
	cfg := config()
	if !cfg.WorkspaceScoped || !cfg.OwnerScoped {
		t.Fatal("ScopeToOwner did not scope the message API")
	}
	other, ws := "u2", "w2"
	got, err := create(t, cfg, &CreateMessageDto{UserId: &other, WorkspaceId: &ws}, ntxctx.NTXContext{UserID: "u1", WorkspaceID: "w1"})
	if err != nil || *got.UserId != "u1" || *got.WorkspaceId != "w1" {
		t.Fatalf("a scoped message was filed for someone else: %+v %v", got, err)
	}
	if _, err := create(t, cfg, &CreateMessageDto{}, ntxctx.NTXContext{}); err == nil {
		t.Fatal("a scoped create without a caller succeeded")
	}
}
