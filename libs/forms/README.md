# forms

Go/Goose microservice for dynamic form definitions and submissions.

## Overview

`forms` stores form schemas, their fields, form types, and submission logs. Other services use it to define configurable input structures without code changes.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/forms` | Form CRUD |
| GET/POST/PATCH/DELETE | `/formfields` | FormField CRUD |
| GET/POST/PATCH/DELETE | `/formlogs` | FormLog CRUD |
| GET/POST/PATCH/DELETE | `/formtypes` | FormType CRUD |

## Entities

- **Form** — name, type, fields, status
- **FormField** — formId, name, type, validation, order
- **FormLog** — formId, userId, data, status
- **FormType** — name, meta

## Module

```go
import "github.com/thescaffold/gox-apps-forms/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
