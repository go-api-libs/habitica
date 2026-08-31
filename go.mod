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
	github.com/MarkRosemaker/openapi-enrich v0.0.0-20260831110212-9eff5623d945
	github.com/go-api-libs/api v0.0.0-20260827160132-fe8c2393f615
	github.com/go-api-libs/types v0.0.0-20260821232109-0cf45378823e
	github.com/google/uuid v1.6.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/MarkRosemaker/errpath v0.0.0-20260827160129-c0d814ff4bdf // indirect
	github.com/MarkRosemaker/json2yaml v0.0.0-20260827160130-e13d71a9e20f // indirect
	github.com/MarkRosemaker/openapi v0.0.0-20260828173639-2eb9293069d5 // indirect
	github.com/MarkRosemaker/openapi-codegen v0.0.0-20260830132421-e79794384c08 // indirect
	github.com/MarkRosemaker/openapi-compare v0.0.0-20260828220208-79b05e83e791 // indirect
	github.com/MarkRosemaker/openapi-compress v0.0.0-20260829131102-dd1725ee634c // indirect
	github.com/MarkRosemaker/openapi-edit v0.0.0-20260829102119-4703cf7c983c // indirect
	github.com/MarkRosemaker/openapi-flatten v0.0.0-20260829103304-14073604f302 // indirect
	github.com/MarkRosemaker/openapi-merge v0.0.0-20260829103842-1c2b2540f717 // indirect
	github.com/MarkRosemaker/ordmap v0.0.0-20260827160235-3615cea69fee // indirect
	github.com/MarkRosemaker/yaml v0.0.0-20260827160238-8e7ad4112fde // indirect
	github.com/MarkRosemaker/yaml2json v0.0.0-20260827160130-ff9effdeb201 // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20260824195058-e88cd73687aa // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)
