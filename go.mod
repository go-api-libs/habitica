module github.com/go-api-libs/habitica

go 1.27

tool (
	github.com/MarkRosemaker/openapi-codegen/cmd/openapi-codegen
	github.com/MarkRosemaker/openapi-compress/cmd/openapi-compress
	github.com/MarkRosemaker/openapi-enrich/cmd/openapi-enrich
	github.com/MarkRosemaker/openapi-flatten/cmd/openapi-flatten
)

require (
	github.com/MarkRosemaker/jsonutil v0.0.0-20260827160132-fe5e496a04a0
	github.com/MarkRosemaker/openapi-enrich v0.0.0-20260823141700-b724bcc12500
	github.com/go-api-libs/api v0.0.0-20260827160132-fe8c2393f615
	github.com/go-api-libs/types v0.0.0-20260821232109-0cf45378823e
	github.com/google/uuid v1.6.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/MarkRosemaker/errpath v0.0.0-20260425165607-bbd4959d04d9 // indirect
	github.com/MarkRosemaker/json2yaml v0.0.0-20260820194645-20aa3a7082f4 // indirect
	github.com/MarkRosemaker/openapi v0.0.0-20260822160201-4ef929e6dacd // indirect
	github.com/MarkRosemaker/openapi-codegen v0.0.0-20260823134004-622339c4d161 // indirect
	github.com/MarkRosemaker/openapi-compare v0.0.0-20260822160308-d46325b41c60 // indirect
	github.com/MarkRosemaker/openapi-compress v0.0.0-20260822224615-33984efb16a7 // indirect
	github.com/MarkRosemaker/openapi-edit v0.0.0-20260822160309-e064283927a6 // indirect
	github.com/MarkRosemaker/openapi-flatten v0.0.0-20260822213818-299a958d1c34 // indirect
	github.com/MarkRosemaker/openapi-merge v0.0.0-20260822160309-abc859c5db13 // indirect
	github.com/MarkRosemaker/ordmap v0.0.0-20260821225345-9c948bb0ea43 // indirect
	github.com/MarkRosemaker/yaml v0.0.0-20260820194724-a126111ba94f // indirect
	github.com/MarkRosemaker/yaml2json v0.0.0-20260820194543-4c959435803e // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20260820142414-ca536658362e // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)
