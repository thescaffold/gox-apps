package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/thescaffold/gox-apps/libs/blobs/app/file"
	"github.com/thescaffold/gox-apps/libs/blobs/app/page"
)

// AppService implements the blobs init/batch/verify/upload pipeline.
// Mirrors ntx-apps/libs/blobs/src/app.service.ts.
type AppService struct {
	fileEntity *file.FileEntity `inject:""`
	pageEntity *page.PageEntity `inject:""`
}

// GetHello mirrors TS getHello() — "Hello World!".
func (s *AppService) GetHello() string { return "Hello World!" }

// InitInput is the deduped set of file fields accepted by Init.
type InitInput struct {
	UserId      string
	ClientId    string
	WorkspaceId string
	Type        string
	Name        string
	ParentId    *string
	Tags        []string
	Size        int64
	Mime        string
	Status      *string
	Meta        json.RawMessage
}

// InitOutput is the file record fields + computed URL, flattened to match TS
// init() which returns `{...nFile, url}` (file fields at the top level, not
// nested under a "file" key). The embedded *file.File promotes its JSON fields.
type InitOutput struct {
	*file.File
	URL string `json:"url"`
}

// Init finds-or-creates a file by (userId, clientId, workspaceId, type, name).
// Mirrors TS AppService.init().
func (s *AppService) Init(in InitInput) (*InitOutput, error) {
	existing, _ := s.fileEntity.First(
		`user_id = ? AND client_id = ? AND workspace_id = ? AND type = ? AND name = ?`,
		in.UserId, in.ClientId, in.WorkspaceId, in.Type, in.Name,
	)
	var f *file.File
	if existing != nil {
		f = existing
	} else {
		f = &file.File{
			UserId:      in.UserId,
			ClientId:    in.ClientId,
			WorkspaceId: in.WorkspaceId,
			Type:        in.Type,
			Name:        in.Name,
			ParentId:    in.ParentId,
			Tags:        in.Tags,
			Size:        in.Size,
			Mime:        in.Mime,
			Status:      in.Status,
			Meta:        in.Meta,
		}
		if err := s.fileEntity.Insert(f); err != nil {
			return nil, err
		}
	}
	return &InitOutput{File: f, URL: s.fileURL(f.Id)}, nil
}

// BatchInput is one chunked page accepted by Batch.
type BatchInput struct {
	FileId string
	Index  int
	Raw    []byte
}

// Batch find-or-creates each (fileId, index) page with the supplied raw bytes.
// Mirrors TS AppService.batch().
func (s *AppService) Batch(items []BatchInput) ([]page.Page, error) {
	out := make([]page.Page, 0, len(items))
	for _, it := range items {
		existing, _ := s.pageEntity.First(`file_id = ? AND "index" = ?`, it.FileId, it.Index)
		if existing != nil {
			out = append(out, *existing)
			continue
		}
		p := &page.Page{FileId: it.FileId, Index: it.Index, Raw: it.Raw}
		if err := s.pageEntity.Insert(p); err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, nil
}

// VerifyInput identifies the file to verify.
type VerifyInput struct {
	UserId      string
	ClientId    string
	WorkspaceId string
	Type        string
	Name        string
	ParentId    *string
}

// VerifyOutput is the flattened file record + pages count + URL, matching TS
// verify() which returns `{...nFile, pagesCount, url}`.
type VerifyOutput struct {
	*file.File
	PagesCount int64  `json:"pagesCount"`
	URL        string `json:"url"`
}

// Verify returns the file plus how many pages have been uploaded.
// Mirrors TS AppService.verify().
func (s *AppService) Verify(in VerifyInput) (*VerifyOutput, error) {
	f, err := s.fileEntity.First(
		`user_id = ? AND client_id = ? AND workspace_id = ? AND type = ? AND name = ?`,
		in.UserId, in.ClientId, in.WorkspaceId, in.Type, in.Name,
	)
	if err != nil || f == nil {
		return nil, fmt.Errorf("blobs: file not found")
	}
	count, err := s.pageEntity.Count(`file_id = ?`, f.Id)
	if err != nil {
		return nil, err
	}
	return &VerifyOutput{File: f, PagesCount: count, URL: s.fileURL(f.Id)}, nil
}

// UploadOutput bundles the upload result. Mirrors TS upload() which returns
// `{file: {...nFile, url}, pages}` — the nested `file` carries the file fields
// AND the computed url (i.e. the flattened InitOutput shape).
type UploadOutput struct {
	File  *InitOutput `json:"file"`
	Pages []page.Page `json:"pages"`
}

// Upload does Init + Batch (with auto-incremented page indexes) in one call.
// Mirrors TS AppService.upload().
func (s *AppService) Upload(in InitInput, rawPages []string) (*UploadOutput, error) {
	initOut, err := s.Init(in)
	if err != nil {
		return nil, err
	}
	batchInputs := make([]BatchInput, 0, len(rawPages))
	for i, raw := range rawPages {
		decoded, decErr := base64.StdEncoding.DecodeString(raw)
		if decErr != nil {
			return nil, decErr
		}
		batchInputs = append(batchInputs, BatchInput{
			FileId: initOut.File.Id, Index: i, Raw: decoded,
		})
	}
	pages, err := s.Batch(batchInputs)
	if err != nil {
		return nil, err
	}
	return &UploadOutput{File: initOut, Pages: pages}, nil
}

// Download returns the assembled raw bytes of a file (concatenated pages in order).
// Mirrors TS AppService.download() but returns []byte instead of streaming to
// a FastifyReply — callers (controller) wrap the bytes in an HTTP response.
func (s *AppService) Download(fileIDOrName string) ([]byte, string, error) {
	id := fileIDOrName
	if !looksLikeUUID(id) {
		f, _ := s.fileEntity.First(`name = ?`, fileIDOrName)
		if f == nil {
			return nil, "", fmt.Errorf("blobs: file not found")
		}
		id = f.Id
	}
	pages, err := s.pageEntity.Find(0, 0, `file_id = ?`, id)
	if err != nil {
		return nil, "", err
	}
	// pages may arrive ordered by created_at desc per Hydrate default — re-sort by index.
	sortPagesByIndex(pages)
	var out []byte
	for _, p := range pages {
		out = append(out, p.Raw...)
	}
	return out, id, nil
}

func (s *AppService) fileURL(id string) string {
	base := os.Getenv("BLOBS_BASE_URL")
	if base == "" {
		return "/" + id
	}
	return base + "/" + id
}

func looksLikeUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

func sortPagesByIndex(pages []page.Page) {
	for i := 1; i < len(pages); i++ {
		for j := i; j > 0 && pages[j-1].Index > pages[j].Index; j-- {
			pages[j-1], pages[j] = pages[j], pages[j-1]
		}
	}
}
