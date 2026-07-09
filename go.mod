module github.com/go-api-libs/habitica

go 1.26.3

tool (
	github.com/MarkRosemaker/openapi-codegen/cmd/openapi-codegen
	github.com/MarkRosemaker/openapi-compress/cmd/openapi-compress
	github.com/MarkRosemaker/openapi-enrich/cmd/openapi-enrich
	github.com/MarkRosemaker/openapi-flatten/cmd/openapi-flatten
)

require (
	github.com/MarkRosemaker/jsonutil v0.0.0-20260707234246-8bcc7216cf7c
	github.com/MarkRosemaker/openapi-enrich v0.0.0-20260708160446-639fe878732d
	github.com/go-api-libs/api v0.0.0-20260705004954-dad48fbb4ab2
	github.com/go-api-libs/types v0.0.0-20251210072721-82754f56609d
	github.com/google/uuid v1.6.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/MarkRosemaker/errpath v0.0.0-20260425165607-bbd4959d04d9 // indirect
	github.com/MarkRosemaker/json2yaml v0.0.0-20260507220148-d6cc0d01bff0 // indirect
	github.com/MarkRosemaker/openapi v0.0.0-20260708040154-7bf5161da792 // indirect
	github.com/MarkRosemaker/openapi-codegen v0.0.0-20260709041146-6f03abc3d3a7 // indirect
	github.com/MarkRosemaker/openapi-compress v0.0.0-20260708183253-c412bc6ba9c6 // indirect
	github.com/MarkRosemaker/openapi-flatten v0.0.0-20260709040850-2a612ba215ef // indirect
	github.com/MarkRosemaker/openapi-merge v0.0.0-20260708040219-26b7609c01a0 // indirect
	github.com/MarkRosemaker/ordmap v0.0.0-20260611220112-724580dd2bee // indirect
	github.com/MarkRosemaker/yaml v0.0.0-20260508005758-fe21a538b084 // indirect
	github.com/MarkRosemaker/yaml2json v0.0.0-20260507220136-7748efc522b2 // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.10.0 // indirect
)
