module github.com/go-api-libs/habitica

go 1.26.3

tool (
	github.com/MarkRosemaker/openapi-codegen/cmd/openapi-codegen
	github.com/MarkRosemaker/openapi-compress/cmd/openapi-compress
	github.com/MarkRosemaker/openapi-enrich/cmd/openapi-enrich
	github.com/MarkRosemaker/openapi-flatten/cmd/openapi-flatten
)

require (
	github.com/MarkRosemaker/jsonutil v0.0.0-20260718153618-78b5039427a4
	github.com/MarkRosemaker/openapi-enrich v0.0.0-20260718220849-60334c972076
	github.com/go-api-libs/api v0.0.0-20260705004954-dad48fbb4ab2
	github.com/go-api-libs/types v0.0.0-20251210072721-82754f56609d
	github.com/google/uuid v1.6.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/MarkRosemaker/errpath v0.0.0-20260425165607-bbd4959d04d9 // indirect
	github.com/MarkRosemaker/json2yaml v0.0.0-20260507220148-d6cc0d01bff0 // indirect
	github.com/MarkRosemaker/openapi v0.0.0-20260718220251-1bf2cab23c1d // indirect
	github.com/MarkRosemaker/openapi-codegen v0.0.0-20260727101120-3b15878a9f50 // indirect
	github.com/MarkRosemaker/openapi-compress v0.0.0-20260718221534-972a6c005575 // indirect
	github.com/MarkRosemaker/openapi-flatten v0.0.0-20260718221215-0f7b06edb27b // indirect
	github.com/MarkRosemaker/openapi-merge v0.0.0-20260718220308-ae0544f2c955 // indirect
	github.com/MarkRosemaker/ordmap v0.0.0-20260718220113-57637f5c3ff8 // indirect
	github.com/MarkRosemaker/yaml v0.0.0-20260508005758-fe21a538b084 // indirect
	github.com/MarkRosemaker/yaml2json v0.0.0-20260507220136-7748efc522b2 // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20260718201538-764159d718ef // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)
