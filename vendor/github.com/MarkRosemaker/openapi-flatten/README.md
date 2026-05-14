# openapi-flatten

`openapi-flatten` is a Go library that eliminates nesting in [OpenAPI 3.x](https://spec.openapis.org/oas/v3.1.0) specifications. It promotes inline schema definitions, responses, request bodies, and parameters into the top-level `components` section and replaces them with `$ref` references.

## Why flatten?

OpenAPI allows schemas to be defined inline anywhere they are used. While convenient for small specs, deeply nested inline definitions make large specs harder to read, harder to reuse, and harder to generate consistent client code from. Flattening gives every meaningful type a name and a single canonical location.

## Installation

```bash
go get -tool github.com/MarkRosemaker/openapi-flatten/cmd/openapi-flatten
```

or

```bash
go get github.com/MarkRosemaker/openapi-flatten
```

## Usage

```go
import (
    "github.com/MarkRosemaker/openapi"
    flatten "github.com/MarkRosemaker/openapi-flatten"
)

// Load an OpenAPI document (JSON or YAML)
doc, err := openapi.LoadFromDataJSON(jsonBytes)
if err != nil {
    log.Fatal(err)
}

// Flatten all inline definitions
if err := flatten.Document(doc); err != nil {
    log.Fatal(err)
}

// doc now has no nested inline objects — only $ref pointers
```

## What gets flattened

### Schemas

Inline schemas are moved to `components/schemas` when they contain meaningful structure. Simple scalar types (`integer`, `number`, `boolean`, plain `string`) stay inline to keep the spec readable. A schema is moved when it is:

- an **object** with properties
- a **string** or **array of strings** with `enum` values
- an **array of objects**

Schemas inside `allOf` are never moved because they exist solely to compose a larger type.

**Before:**

```json
{
  "paths": {
    "/pets": {
      "post": {
        "responses": {
          "400": {
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": { "type": "string" }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

**After:**

```json
{
  "paths": {
    "/pets": {
      "post": {
        "responses": {
          "400": {
            "$ref": "#/components/responses/CreatePetBadRequestResponse"
          }
        }
      }
    }
  },
  "components": {
    "responses": {
      "CreatePetBadRequestResponse": {
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/CreatePetBadRequestJsonResponse"
            }
          }
        }
      }
    },
    "schemas": {
      "CreatePetBadRequestJsonResponse": {
        "type": "object",
        "properties": {
          "error": { "type": "string" }
        }
      }
    }
  }
}
```

### Responses

Every inline response object is moved to `components/responses`. The generated name combines the operation ID, the HTTP status text, and the suffix `Response`:

```
{OperationID}{StatusText}Response
```

Examples: `CreatePetBadRequestResponse`, `GetMeUnauthorizedResponse`.

Error responses (status ≥ 400) always have their schemas promoted to components. Success responses only promote complex schemas.

### Request bodies

Inline request bodies are moved to `components/requestBodies`. The generated name is:

```
{OperationID}RequestBody
```

Example: `CreatePetRequestBody`.

### Parameters

Inline parameters are moved to `components/parameters` using the parameter's own `name` field.

## Name generation

All names are converted to Go-style PascalCase (e.g., `create pet bad request response` → `CreatePetBadRequestResponse`). If the generated name is already taken, a numeric suffix is appended (`Name2`, `Name3`, …) to avoid collisions.

## Error reporting

Errors include the full JSON path to the offending field, powered by [`errpath`](https://github.com/MarkRosemaker/errpath):

```
paths["/pets"].post.responses["400"]["application/json"].schema: unimplemented schema ref type "null"
```

## Dependencies

| Package | Purpose |
|---|---|
| [`github.com/MarkRosemaker/openapi`](https://github.com/MarkRosemaker/openapi) | OpenAPI 3.x data structures |
| [`github.com/MarkRosemaker/errpath`](https://github.com/MarkRosemaker/errpath) | Error path context |
| [`github.com/ettle/strcase`](https://github.com/ettle/strcase) | PascalCase name conversion |

## License

See [LICENSE](LICENSE).
