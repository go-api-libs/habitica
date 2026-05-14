package codegen

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/MarkRosemaker/openapi-codegen/ir"
	"github.com/MarkRosemaker/openapi-enrich/cassette"
)

// matchInteractions populates doc.InteractionCalls by matching each 2xx
// interaction to a known operation and extracting Go literal argument values.
func matchInteractions(doc *ir.Document, interactions cassette.Interactions) error {
	for _, ia := range interactions {
		u, err := url.Parse(ia.Request.URL)
		if err != nil {
			return err
		}

		// Strip the base URL path prefix to get the operation-relative path.
		relPath := strings.TrimPrefix(u.Path, doc.BaseURL.Path)
		if relPath == "" {
			relPath = "/"
		} else if !strings.HasPrefix(relPath, "/") {
			relPath = "/" + relPath
		}

		for i := range doc.Operations {
			op := &doc.Operations[i]
			if op.Method != ia.Request.Method {
				continue
			}

			pathVals, ok := matchPathTemplate(op.PathTemplate, relPath)
			if !ok {
				continue
			}

			call := ir.InteractionCall{Op: op}

			for _, pp := range op.PathParams {
				call.PathArgs = append(call.PathArgs, goLiteralForType(pp.Type, pathVals[pp.JSONName]))
			}

			q := u.Query()
			for _, qp := range op.QueryParams {
				val := q.Get(qp.JSONName)
				if val != "" {
					call.QueryArgs = append(call.QueryArgs, ir.InteractionParam{
						FieldName: qp.FieldName,
						Literal:   goLiteralForType(qp.Type, val),
					})
				}
			}

			doc.InteractionCalls = append(doc.InteractionCalls, call)
			break
		}
	}

	return nil
}

// matchPathTemplate matches a URL path against an operation path template,
// returning the extracted parameter values keyed by parameter name.
// Template segments may be full-segment ({name}) or mid-segment (CIK{name}.json).
func matchPathTemplate(template, path string) (map[string]string, bool) {
	tParts := strings.Split(template, "/")
	pParts := strings.Split(path, "/")

	if len(tParts) != len(pParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, tp := range tParts {
		if !extractSegmentParam(tp, pParts[i], params) {
			return nil, false
		}
	}

	return params, true
}

// extractSegmentParam matches one template segment against one path segment,
// populating out with any extracted parameter values. Returns false on mismatch.
func extractSegmentParam(tmpl, value string, out map[string]string) bool {
	if !strings.Contains(tmpl, "{") {
		return tmpl == value
	}

	// Simple case: entire segment is {name}
	if strings.HasPrefix(tmpl, "{") && strings.HasSuffix(tmpl, "}") && strings.Count(tmpl, "{") == 1 {
		out[tmpl[1:len(tmpl)-1]] = value
		return true
	}

	// Mid-segment case: e.g. "CIK{name}.json"
	start := strings.Index(tmpl, "{")
	end := strings.Index(tmpl, "}")
	if start < 0 || end <= start {
		return tmpl == value
	}

	prefix := tmpl[:start]
	suffix := tmpl[end+1:]
	name := tmpl[start+1 : end]

	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, suffix) {
		return false
	}

	extracted := value[len(prefix):]
	if len(suffix) > 0 {
		extracted = extracted[:len(extracted)-len(suffix)]
	}
	out[name] = extracted
	return true
}

// goLiteralForType returns a Go literal expression for value given the Go type.
func goLiteralForType(goType, value string) string {
	switch goType {
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return value
	case "uuid.UUID":
		return fmt.Sprintf("uuid.MustParse(%q)", value)
	default:
		return fmt.Sprintf("%q", value)
	}
}
