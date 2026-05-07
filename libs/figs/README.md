# figs

Go/Goose microservice for file conversion, mapping, validation, and storage.

## Overview

`figs` is a document-processing pipeline. It accepts named file payloads, converts them (CSV, Excel, PDF), validates their schema (JSON/YAML), maps their structure (Data, HTML), and stores the result (locally or in S3).

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/file` | CRUD for File records |
| POST | `/:name` | Run the SaveFile pipeline for a named file definition |

## Entities

- **File** — name, input, output, meta, url, raw, tags, status

## Packages

| Package | Implementations |
|---------|----------------|
| `pkg/converter` | CSV (real), Excel (excelize), PDF (fpdf) |
| `pkg/mapper` | Data (JSON transform), HTML (template render) |
| `pkg/store` | Local (filesystem), Object (S3-compatible via AWS SDK v2) |
| `pkg/validator` | JSON Schema (gojsonschema), YAML Schema (gojsonschema + yaml.v3) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FIGS_S3_BUCKET` | — | S3 bucket for object store |
| `FIGS_S3_REGION` | `us-east-1` | AWS region |
| `FIGS_S3_ENDPOINT` | — | Custom endpoint (MinIO, etc.) |
| `AWS_ACCESS_KEY_ID` | — | Static AWS credentials |
| `AWS_SECRET_ACCESS_KEY` | — | Static AWS credentials |

## Module

```go
import "github.com/thescaffold/gox-apps-figs/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
