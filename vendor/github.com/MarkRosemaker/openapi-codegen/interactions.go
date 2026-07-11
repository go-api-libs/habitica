package codegen

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/MarkRosemaker/openapi-codegen/ir"
	"github.com/MarkRosemaker/openapi-enrich/cassette"
)

// matchInteractions populates doc.InteractionCalls by matching each interaction
// to a known operation and extracting Go literal argument values.
// Every interaction must match exactly one operation; an error is returned otherwise,
// guaranteeing len(doc.InteractionCalls) == len(interactions) on success.
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

		op, pathVals := pickOperation(doc.Operations, ia.Request.Method, relPath)
		if op == nil {
			return fmt.Errorf("interaction %s %s: no matching operation found", ia.Request.Method, ia.Request.URL)
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

		for _, hp := range op.HeaderParams {
			val := ia.Request.Headers.Get(hp.JSONName)
			if val != "" {
				call.HeaderArgs = append(call.HeaderArgs, ir.InteractionParam{
					FieldName: hp.FieldName,
					Literal:   goLiteralForType(hp.Type, val),
				})
			}
		}

		doc.InteractionCalls = append(doc.InteractionCalls, call)
	}

	return nil
}

// pickOperation returns the operation whose method and path template match the
// URL, preferring an exact segment-count match over a trailing-wildcard match
// so that e.g. /tasks/{taskId}/score/up wins over /tasks/{taskId} for a URL
// like /tasks/abc/score/up.
func pickOperation(ops []ir.Operation, method, relPath string) (*ir.Operation, map[string]string) {
	var wildcardOp *ir.Operation
	var wildcardVals map[string]string
	for i := range ops {
		op := &ops[i]
		if op.Method != method {
			continue
		}

		vals, ok, exact := matchPathTemplate(op.PathTemplate, relPath)
		if !ok {
			continue
		}

		if exact {
			return op, vals
		}

		if wildcardOp == nil {
			wildcardOp = op
			wildcardVals = vals
		}
	}

	return wildcardOp, wildcardVals
}

// matchPathTemplate matches a URL path against an operation path template,
// returning the extracted parameter values keyed by parameter name.
// Template segments may be full-segment ({name}) or mid-segment (CIK{name}.json).
// A trailing pure {param} segment acts as a wildcard that consumes all remaining
// path segments joined by "/", enabling multi-segment captures like {path} in
// /package/{path} matching /package/github.com/user/repo. The third return
// value reports whether the match was segment-for-segment exact; callers should
// prefer an exact match over a wildcard one.
func matchPathTemplate(template, path string) (params map[string]string, matched, exact bool) {
	tParts := strings.Split(template, "/")
	pParts := strings.Split(path, "/")

	if len(tParts) > len(pParts) {
		return nil, false, false
	}

	params = make(map[string]string)
	for i, tp := range tParts {
		// Last template segment that is a pure {param} and there are extra path
		// segments: consume all remaining segments as one slash-joined value.
		if i == len(tParts)-1 &&
			strings.HasPrefix(tp, "{") &&
			strings.HasSuffix(tp, "}") &&
			strings.Count(tp, "{") == 1 &&
			len(pParts) > i+1 {
			params[tp[1:len(tp)-1]] = strings.Join(pParts[i:], "/")
			return params, true, false
		}

		if !extractSegmentParam(tp, pParts[i], params) {
			return nil, false, false
		}
	}

	sameLen := len(tParts) == len(pParts)
	return params, sameLen, sameLen
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
