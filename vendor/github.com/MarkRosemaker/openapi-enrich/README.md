# openapi-enrich

Enriches an OpenAPI 3.1 document from observed HTTP traffic.

## Overview

`openapi-enrich` is a focused library with the main public function:

```go
func Enrich(doc *openapi.Document, interactions cassette.Interactions) error
```

Feed it an OpenAPI document and a slice of observed HTTP interactions; it adds
paths, operations, parameters, request bodies, and response schemas inferred
from the traffic.  Schemas are left inline — the caller composes any
post-processing (flatten, tidy, sort) as needed.

## Install

```bash
go get -tool github.com/MarkRosemaker/openapi-enrich/cmd/openapi-enrich
```

## Usage

```go
import (
    enrich "github.com/MarkRosemaker/openapi-enrich"
    "github.com/MarkRosemaker/openapi-enrich/cassette"
)

// Start from a minimal document or load an existing spec.
doc := enrich.NewDocument()

interactions := []cassette.Interaction{
    {
        Request: cassette.Request{
            Method:  "GET",
            URL:     "https://api.example.com/users",
            Headers: http.Header{},
        },
        Response: cassette.Response{
            StatusCode: http.StatusOK,
            Headers:    http.Header{"Content-Type": {"application/json"}},
            Body:       []byte(`[{"id":1,"name":"Alice"}]`),
        },
    },
}

if err := enrich.Enrich(doc, interactions); err != nil {
    log.Fatal(err)
}
```

## What it infers

- **Paths** — detected from request URLs, with ID-like segments replaced by
  `{param}` path parameters.
- **Operations** — one per unique method + path, with an inferred `operationId`
  (e.g. `GET /users` → `ListUsers`, `GET /users/{id}` → `GetUserByID`).
- **Query parameters** — schema inferred from values; comma-separated values
  become non-exploded arrays.
- **Request headers** — `Authorization` creates an HTTP security scheme;
  `x-*` and other custom headers become header parameters.
- **Request bodies** — JSON bodies produce inline object schemas.
- **Responses** — JSON, text/plain, and text/html responses are modeled;
  repeated observations are merged.
- **Schema formats** — UUID, URI, email, date-time, IPv4, IPv6 are detected
  automatically from string values.

## Design

- **No I/O** — the caller loads and saves the spec.
- **No flatten/tidy/sort** — use separate libraries for those.
- **Own interaction types** — no dependency on a specific HTTP recording format.

## Requirements

Requires Go 1.25 with `GOEXPERIMENT=jsonv2` (set via `go env -w GOEXPERIMENT=jsonv2`).
