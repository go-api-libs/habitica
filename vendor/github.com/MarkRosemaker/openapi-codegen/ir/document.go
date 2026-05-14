package ir

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/MarkRosemaker/openapi"
	compress "github.com/MarkRosemaker/openapi-compress"
	flatten "github.com/MarkRosemaker/openapi-flatten"
)

// FromDocument converts a fully-loaded and flattened openapi.Document to an IR Document.
// cfg provides the package name and optional user-agent override.
func FromDocument(doc *openapi.Document, packageName, userAgent string) (*Document, error) {
	if err := flatten.Document(doc); err != nil {
		return nil, fmt.Errorf("flatten: %w", err)
	}

	if err := compress.Document(doc, compress.Config{}); err != nil {
		return nil, fmt.Errorf("compress: %w", err)
	}

	// sort the components and responses
	// (but not the paths since it may be good to keep them in the order they were added)
	doc.Components.SortMaps()
	for _, path := range doc.Paths {
		for _, op := range path.Operations {
			op.Responses.Sort()
		}
	}

	baseURL, err := parseBaseURL(doc)
	if err != nil {
		return nil, fmt.Errorf("servers: %w", err)
	}

	if userAgent == "" && doc.Info != nil {
		userAgent = doc.Info.Title
	}

	schemas, err := FromComponentSchemas(doc.Components.Schemas)
	if err != nil {
		return nil, fmt.Errorf("components.schemas: %w", err)
	}

	operations, err := fromPaths(doc.Paths)
	if err != nil {
		return nil, fmt.Errorf("paths: %w", err)
	}

	security := fromSecurity(doc.Components.SecuritySchemes)

	hasURL, hasDuration, hasDate := needsSpecialImports(schemas, operations)

	return &Document{
		PackageName:       packageName,
		BaseURL:           baseURL,
		UserAgent:         userAgent,
		Schemas:           schemas,
		Operations:        operations,
		Security:          security,
		HasURLFields:      hasURL,
		HasDurationFields: hasDuration,
		HasDateFields:     hasDate,
	}, nil
}

// parseBaseURL extracts scheme, host, and path from the first server URL.
func parseBaseURL(doc *openapi.Document) (URLParts, error) {
	if len(doc.Servers) == 0 {
		return URLParts{Scheme: "https"}, nil
	}
	u, err := url.Parse(doc.Servers[0].URL)
	if err != nil {
		return URLParts{}, fmt.Errorf("parse %q: %w", doc.Servers[0].URL, err)
	}
	return URLParts{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   strings.TrimSuffix(u.Path, "/"),
	}, nil
}

// fromPaths iterates all path items and operations, converting each to ir.Operation.
func fromPaths(paths openapi.Paths) ([]Operation, error) {
	var ops []Operation
	for path, item := range paths.ByIndex() {
		for method, op := range item.Operations {
			irOp, err := FromOperation(path, item.Parameters, method, op)
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", method, path, err)
			}
			ops = append(ops, *irOp)
		}
	}
	return ops, nil
}

func fromSecurity(schemes openapi.SecuritySchemes) Security {
	s := Security{}

	for _, sec := range schemes {
		v := sec.Value
		switch v.Scheme {
		case openapi.SecuritySchemeBearer:
			s.Bearer = Bearer{
				Name: v.Name,
			}
		}
	}

	return s
}

// needsSpecialImports scans schemas and operation types for url.URL, time.Duration, civil.Date.
func needsSpecialImports(schemas []Schema, ops []Operation) (hasURL, hasDuration, hasDate bool) {
	check := func(goType string) {
		if containsType(goType, "url.URL") {
			hasURL = true
		}
		if containsType(goType, "time.Duration") {
			hasDuration = true
		}
		if containsType(goType, "civil.Date") {
			hasDate = true
		}
	}

	for _, s := range schemas {
		for _, f := range s.Fields {
			check(f.Type)
		}
	}

	for _, op := range ops {
		for _, p := range op.PathParams {
			check(p.Type)
		}
		for _, p := range op.QueryParams {
			check(p.Type)
		}
		if op.SuccessReturn != nil {
			check(op.SuccessReturn.Name)
		}
	}

	return hasURL, hasDuration, hasDate
}

func containsType(goType, needle string) bool {
	return goType == needle || strings.Contains(goType, needle)
}
