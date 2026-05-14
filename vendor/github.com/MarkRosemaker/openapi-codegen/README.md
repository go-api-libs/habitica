# openapi-codegen

Parses an OpenAPI 3.x spec, flattens it, and generates idiomatic Go code — types, HTTP client, HTTP server, and tests.

## Install

```bash
go get -tool github.com/MarkRosemaker/openapi-codegen/cmd/openapi-codegen
```

## Usage

```sh
openapi-codegen -spec openapi.json -out ./gen -package mypkg
```

## Generated output

- **types** — structs, enums, and type aliases for all referenced schemas
- **client** — typed HTTP client with per-operation methods
- **server** — `http.Handler`-based server scaffold
- **tests** — round-trip and VCR-cassette tests

## Dependencies of generated code

| Module | Purpose |
|---|---|
| `github.com/go-api-libs/api` | `ErrUnknownStatusCode`, `WrapDecodingError` |
| `github.com/go-api-libs/types` | `types.Email` etc. |
| `github.com/MarkRosemaker/jsonutil` | JSON marshalers for `url.URL`, `time.Duration` |
| `github.com/google/uuid` | `uuid.UUID` (when spec uses UUID format) |
| `cloud.google.com/go/civil` | `civil.Date` (when spec uses date format) |

## License

MIT
