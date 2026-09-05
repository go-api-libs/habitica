module github.com/go-api-libs/habitica

go 1.27

tool (
	github.com/MarkRosemaker/openapi-codegen/cmd/openapi-codegen
	github.com/MarkRosemaker/openapi-compress/cmd/openapi-compress
	github.com/MarkRosemaker/openapi-enrich/cmd/openapi-enrich
	github.com/MarkRosemaker/openapi-flatten/cmd/openapi-flatten
)

require (
	github.com/MarkRosemaker/jsonutil v0.0.0-20260905160105-5e81f9654374
	github.com/MarkRosemaker/openapi-enrich v0.0.0-20260905161434-377d6441ee4e
	github.com/go-api-libs/api v0.0.0-20260905160107-5bf91bbd722f
	github.com/go-api-libs/types v0.0.0-20260821232109-0cf45378823e
	github.com/google/uuid v1.6.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/MarkRosemaker/errpath v0.0.0-20260905160101-b767f06676ce // indirect
	github.com/MarkRosemaker/json2yaml v0.0.0-20260905160102-90094cbdc27a // indirect
	github.com/MarkRosemaker/openapi v0.0.0-20260905160141-933af336d1e3 // indirect
	github.com/MarkRosemaker/openapi-codegen v0.0.0-20260905162804-6514c7779213 // indirect
	github.com/MarkRosemaker/openapi-compare v0.0.0-20260905160218-f0ea74662d7d // indirect
	github.com/MarkRosemaker/openapi-compress v0.0.0-20260905162311-36f05ed3e07c // indirect
	github.com/MarkRosemaker/openapi-edit v0.0.0-20260905160213-02eab925b41a // indirect
	github.com/MarkRosemaker/openapi-flatten v0.0.0-20260905160249-6aed8b0b43cd // indirect
	github.com/MarkRosemaker/openapi-merge v0.0.0-20260905160215-6d4259900033 // indirect
	github.com/MarkRosemaker/ordmap v0.0.0-20260905160117-d1940fc06596 // indirect
	github.com/MarkRosemaker/yaml v0.0.0-20260905160118-287c5415b0bb // indirect
	github.com/MarkRosemaker/yaml2json v0.0.0-20260905160101-04874ebebfcf // indirect
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
