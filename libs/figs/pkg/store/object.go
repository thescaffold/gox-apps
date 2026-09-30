package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/thescaffold/gox-apps/libs/figs/pkg/converter"
	"github.com/thescaffold/gox-packages/libs/blobs/objectstore"
)

// ObjectProvider stores converted files through the ObjectStore (PLAN
// M1-05a; it used to call S3 directly). The key is
// ws/<workspaceId>/figs/<name>.<ext> — the tenant prefix of the shared key
// layout — and the returned URL is that KEY: a presigned link would expire,
// so callers mint one with ObjectStore.Presign when they need to serve it.
type ObjectProvider struct {
	Objects objectstore.ObjectStore
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._ -]+`)

// segment reduces caller text to one safe key segment ("" if nothing is left).
func segment(s string) string {
	s = unsafeName.ReplaceAllString(s, "_")
	s = strings.Trim(s, " .")
	return s
}

func (p *ObjectProvider) Store(payload *Payload, raw *converter.Response) (*Response, error) {
	if p.Objects == nil {
		return nil, errors.New("store: no object store configured (inject one into store.Service; BLOBS_BACKEND selects the driver)")
	}
	name := segment(fmt.Sprint(payload.Meta["name"]))
	if name == "" || name == "<nil>" {
		return nil, errors.New("store: meta.name is required for the object store")
	}
	ws, _ := payload.Meta["workspaceId"].(string)
	if ws = segment(ws); ws == "" {
		ws = "shared"
	}
	ext := segment(raw.Extension)
	if ext != "" {
		name += "." + ext
	}
	key := "ws/" + ws + "/figs/" + name
	_, err := p.Objects.Put(context.Background(), key, bytes.NewReader(raw.Buffer), objectstore.PutOptions{MediaType: raw.Mime})
	if err != nil {
		return nil, fmt.Errorf("store: object store put: %w", err)
	}
	return &Response{URL: key, Raw: string(raw.Buffer)}, nil
}
