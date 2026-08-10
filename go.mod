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
	github.com/MarkRosemaker/openapi-enrich v0.0.0-20260810160723-1cdca9c8073d
	github.com/go-api-libs/api v0.0.0-20260705004954-dad48fbb4ab2
	github.com/go-api-libs/types v0.0.0-20251210072721-82754f56609d
	github.com/google/uuid v1.6.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/MarkRosemaker/errpath v0.0.0-20260425165607-bbd4959d04d9 // indirect
	github.com/MarkRosemaker/json2yaml v0.0.0-20260507220148-d6cc0d01bff0 // indirect
	github.com/MarkRosemaker/openapi v0.0.0-20260810160207-5a25126e941b // indirect
	github.com/MarkRosemaker/openapi-codegen v0.0.0-20260810221031-0fe4b32489fb // indirect
	github.com/MarkRosemaker/openapi-compress v0.0.0-20260810161306-1989090a73cc // indirect
	github.com/MarkRosemaker/openapi-flatten v0.0.0-20260810161015-ab386a864426 // indirect
	github.com/MarkRosemaker/openapi-merge v0.0.0-20260810160223-3e9372f0ae2b // indirect
	github.com/MarkRosemaker/ordmap v0.0.0-20260810160032-e8939f404a04 // indirect
	github.com/MarkRosemaker/yaml v0.0.0-20260508005758-fe21a538b084 // indirect
	github.com/MarkRosemaker/yaml2json v0.0.0-20260507220136-7748efc522b2 // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20260810151157-a8b543ca52da // indirect
	golang.org/x/mod v0.39.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)
