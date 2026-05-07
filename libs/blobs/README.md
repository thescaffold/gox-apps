# blobs

Go/Goose microservice for binary large object (blob) storage and page management.

## Overview

`blobs` manages file uploads and multi-page documents. Files are stored with metadata (userId, clientId, workspaceId, type, tags, mime, size) and can be split into indexed Page records.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/file` | CRUD for File records |
| GET/POST/PATCH/DELETE | `/page` | CRUD for Page records |

## Entities

- **File** — userId, clientId, workspaceId, type, parentId, name, tags, size, mime, status, meta
- **Page** — fileId, index, raw content, status, meta

## Module

```go
import "github.com/thescaffold/gox-apps-blobs/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
