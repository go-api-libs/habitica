module github.com/go-api-libs/habitica

go 1.27

tool (
	github.com/MarkRosemaker/openapi-codegen/cmd/openapi-codegen
	github.com/MarkRosemaker/openapi-compress/cmd/openapi-compress
	github.com/MarkRosemaker/openapi-enrich/cmd/openapi-enrich
	github.com/MarkRosemaker/openapi-flatten/cmd/openapi-flatten
)

require (
	github.com/MarkRosemaker/jsonutil v0.0.0-20260820212410-12ba6685df41
	github.com/MarkRosemaker/openapi-enrich v0.0.0-20260811160812-77c2915f3f46
	github.com/go-api-libs/api v0.0.0-20260705004954-dad48fbb4ab2
	github.com/go-api-libs/types v0.0.0-20251210072721-82754f56609d
	github.com/google/uuid v1.6.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/MarkRosemaker/errpath v0.0.0-20260425165607-bbd4959d04d9 // indirect
	github.com/MarkRosemaker/json2yaml v0.0.0-20260507220148-d6cc0d01bff0 // indirect
	github.com/MarkRosemaker/openapi v0.0.0-20260811160236-f3751aa1c44d // indirect
	github.com/MarkRosemaker/openapi-codegen v0.0.0-20260811221453-406d9247d275 // indirect
	github.com/MarkRosemaker/openapi-compress v0.0.0-20260811221234-d691b2cdf6af // indirect
	github.com/MarkRosemaker/openapi-flatten v0.0.0-20260811220926-763643060982 // indirect
	github.com/MarkRosemaker/openapi-merge v0.0.0-20260811160254-c8256ee94870 // indirect
	github.com/MarkRosemaker/ordmap v0.0.0-20260811160034-24bdeb1fbeab // indirect
	github.com/MarkRosemaker/yaml v0.0.0-20260508005758-fe21a538b084 // indirect
	github.com/MarkRosemaker/yaml2json v0.0.0-20260507220136-7748efc522b2 // indirect
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20260811152304-ee035b5b010f // indirect
	golang.org/x/mod v0.39.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)
