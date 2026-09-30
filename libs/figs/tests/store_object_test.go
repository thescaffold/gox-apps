package tests

import (
	"errors"
	"strings"
	"testing"

	"github.com/thescaffold/gox-apps/libs/figs/pkg/converter"
	"github.com/thescaffold/gox-apps/libs/figs/pkg/store"
	"github.com/thescaffold/gox-packages/libs/blobs/objectstore"
	"github.com/thescaffold/gox-packages/libs/blobs/objectstore/fs"
)

// PLAN M1-05a: figs' `object` store delegates to the ObjectStore (Postgres by
// default) instead of talking to S3 directly.

func objStore(t *testing.T) objectstore.ObjectStore {
	t.Helper()
	s, err := fs.New(t.TempDir(), fs.Config{PresignSecret: []byte("0123456789abcdef0123")})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestObjectProvider_StoresThroughTheObjectStore(t *testing.T) {
	st := objStore(t)
	svc := &store.Service{Objects: st}
	raw := &converter.Response{Buffer: []byte("a,b\n1,2\n"), Extension: "csv", Mime: "text/csv"}
	resp, err := svc.Use(store.Object).Store(&store.Payload{Meta: map[string]any{"name": "report", "workspaceId": "wsA"}}, raw)
	if err != nil {
		t.Fatal(err)
	}
	if resp.URL != "ws/wsA/figs/report.csv" {
		t.Fatalf("URL = %q, want the object key (links are minted on read; a presigned URL would expire)", resp.URL)
	}
	rc, info, err := st.Get(t.Context(), resp.URL, nil)
	if err != nil {
		t.Fatalf("the object is not in the store: %v", err)
	}
	defer func() { _ = rc.Close() }()
	if info.MediaType != "text/csv" || info.Size != int64(len(raw.Buffer)) {
		t.Fatalf("%+v", info)
	}
}

func TestObjectProvider_KeysAreTenantScopedAndSafe(t *testing.T) {
	st := objStore(t)
	svc := &store.Service{Objects: st}
	raw := &converter.Response{Buffer: []byte("x"), Extension: "csv", Mime: "text/csv"}
	for _, name := range []string{"../../escape", "a/b", "..", "", "with space"} {
		resp, err := svc.Use(store.Object).Store(&store.Payload{Meta: map[string]any{"name": name, "workspaceId": "wsA"}}, raw)
		if err != nil {
			continue // rejected outright is fine
		}
		if !strings.HasPrefix(resp.URL, "ws/wsA/figs/") || objectstore.ValidateKey(resp.URL) != nil {
			t.Errorf("name %q produced key %q outside the tenant prefix", name, resp.URL)
		}
		if strings.Count(resp.URL, "/") != 3 {
			t.Errorf("name %q produced a multi-segment key %q (a name must stay one segment)", name, resp.URL)
		}
	}
	// A hostile workspace id cannot climb out either.
	if resp, err := svc.Use(store.Object).Store(&store.Payload{Meta: map[string]any{"name": "n", "workspaceId": "../wsB"}}, raw); err == nil && !strings.HasPrefix(resp.URL, "ws/") {
		t.Errorf("hostile workspace id produced %q", resp.URL)
	}
}

func TestObjectProvider_WithoutAStoreFailsClearly(t *testing.T) {
	_, err := (&store.Service{}).Use(store.Object).Store(&store.Payload{Meta: map[string]any{"name": "n"}}, &converter.Response{Buffer: []byte("x"), Extension: "csv"})
	if err == nil || !strings.Contains(err.Error(), "object store") {
		t.Fatalf("got %v", err)
	}
	_ = errors.Is
}
