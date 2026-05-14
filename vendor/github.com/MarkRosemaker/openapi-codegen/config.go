package codegen

import (
	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-enrich/cassette"
	"github.com/spf13/afero"
)

// Config holds the parameters for a single code-generation run.
type Config struct {
	SpecPath         string                // path to the OpenAPI spec file (JSON or YAML)
	Spec             *openapi.Document     // OpenAPI spec, if already parsed
	InteractionsPath string                // path to the HTTP interactions with the API
	Interactions     cassette.Interactions // sample HTTP interactions with the API
	OutputDir        string                // directory to write the generated Go files
	OutputFs         afero.Fs              // Filesystem to write the generated Go files
	PackageName      string                // Go package name for the generated code
	UserAgent        string                // User-Agent header value for the generated client

	// File selection: if all are false, every file is generated.
	// Set one or more to true to generate only those files.
	Types      bool // generate types.gen.go
	Client     bool // generate client.gen.go
	ClientTest bool // generate client.gen_test.go
	Server     bool // generate server.gen.go
}

// shouldGenerate reports whether filename f should be written given cfg.
func (cfg Config) shouldGenerate(f string) bool {
	if !cfg.Types && !cfg.Client && !cfg.ClientTest && !cfg.Server {
		return true // default: generate everything
	}

	switch f {
	case "types.gen.go":
		return cfg.Types
	case "client.gen.go":
		return cfg.Client
	case "client.gen_test.go":
		return cfg.ClientTest
	case "server.gen.go":
		return cfg.Server
	}

	return true
}
