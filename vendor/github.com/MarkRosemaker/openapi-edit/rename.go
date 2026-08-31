// Package edit provides structural edits to an OpenAPI document — changes
// where touching one place obliges you to touch several others.
package edit

import (
	"fmt"
	"regexp"

	"github.com/MarkRosemaker/openapi"
)

// schemaRefPrefix is the start of every reference to a component schema.
const schemaRefPrefix = "#/components/schemas/"

// reComponentKey is the pattern the OpenAPI specification requires of keys
// under components.
//
// It matters here beyond validity: a name containing "/" would produce a
// reference that resolves somewhere else entirely, and one containing a space
// would produce a reference that does not resolve at all.
//
// See https://spec.openapis.org/oas/v3.1.0#components-object
var reComponentKey = regexp.MustCompile(`^[a-zA-Z0-9.\-_]+$`)

// ErrSchemaNotFound is returned when the schema to rename is not in
// components.schemas.
type ErrSchemaNotFound struct{ Name string }

func (e *ErrSchemaNotFound) Error() string {
	return fmt.Sprintf("schema %q not found in components.schemas", e.Name)
}

// ErrSchemaExists is returned when the new name is already taken by another
// schema. Renaming onto it would silently merge two definitions into one.
type ErrSchemaExists struct{ Name string }

func (e *ErrSchemaExists) Error() string {
	return fmt.Sprintf("schema %q already exists in components.schemas", e.Name)
}

// ErrInvalidSchemaName is returned when the new name is not a valid key under
// components, and so could not be referenced.
type ErrInvalidSchemaName struct{ Name string }

func (e *ErrInvalidSchemaName) Error() string {
	return fmt.Sprintf("schema name %q must match %s", e.Name, reComponentKey)
}

// RenameSchema renames a schema in components.schemas and rewrites every
// reference to it, wherever in the document that reference occurs.
//
// The schema keeps its position among the components, so renaming produces a
// one-line change rather than reordering the section.
//
// Renaming a schema to its current name does nothing and reports no error.
// Otherwise it fails, changing nothing, if:
//
//   - oldName is not in components.schemas ([ErrSchemaNotFound]);
//   - newName is already taken ([ErrSchemaExists]) — renaming onto an existing
//     schema would silently discard one of two different definitions, and
//     point every reference to whichever survived;
//   - newName could not be referenced ([ErrInvalidSchemaName]).
func RenameSchema(doc *openapi.Document, oldName, newName string) error {
	schemas := doc.Components.Schemas

	s, ok := schemas[oldName]
	if !ok {
		return &ErrSchemaNotFound{Name: oldName}
	}

	if oldName == newName {
		return nil
	}

	if !reComponentKey.MatchString(newName) {
		return &ErrInvalidSchemaName{Name: newName}
	}

	if _, exists := schemas[newName]; exists {
		return &ErrSchemaExists{Name: newName}
	}

	// Assign directly rather than through Set: the schema carries its own
	// ordering index, so moving the value to a new key keeps it where it was,
	// while Set would move it to the end of the section.
	delete(schemas, oldName)
	schemas[newName] = s

	renameRefs(doc, schemaRefPrefix+oldName, schemaRefPrefix+newName)

	return nil
}

// renameRefs rewrites every schema reference in doc from old to new.
//
// Schemas can refer to one another in a cycle, so visited records the schemas
// already walked. It holds schemas rather than references because the same
// schema is reachable through several references.
func renameRefs(doc *openapi.Document, old, new string) {
	w := &refWriter{old: old, new: new, visited: map[*openapi.Schema]bool{}}

	for _, s := range doc.Components.Schemas {
		w.schema(s)
	}

	for _, r := range doc.Components.Responses {
		w.response(r)
	}

	for _, p := range doc.Components.Parameters {
		w.parameter(p)
	}

	for _, rb := range doc.Components.RequestBodies {
		w.requestBody(rb)
	}

	w.headers(doc.Components.Headers)

	for _, c := range doc.Components.Callbacks {
		w.callbackRef(c)
	}

	for _, p := range doc.Components.PathItems {
		w.pathItemRef(p)
	}

	for _, p := range doc.Paths {
		w.pathItem(p)
	}

	for _, p := range doc.Webhooks {
		w.pathItemRef(p)
	}
}

// refWriter rewrites one reference identifier throughout a document.
type refWriter struct {
	old, new string
	visited  map[*openapi.Schema]bool
}

func (w *refWriter) pathItemRef(r *openapi.PathItemRef) {
	if r != nil {
		w.pathItem(r.Value)
	}
}

func (w *refWriter) pathItem(p *openapi.PathItem) {
	if p == nil {
		return
	}

	w.parameterList(p.Parameters)

	for _, op := range p.Operations {
		w.operation(op)
	}
}

func (w *refWriter) operation(op *openapi.Operation) {
	if op == nil {
		return
	}

	w.parameterList(op.Parameters)
	w.requestBody(op.RequestBody)

	for _, r := range op.Responses {
		w.response(r)
	}

	for _, c := range op.Callbacks {
		w.callback(c)
	}
}

// callbackRef covers components.callbacks, which holds references, whereas an
// operation holds callbacks by value.
func (w *refWriter) callbackRef(r *openapi.CallbackRef) {
	if r != nil && r.Value != nil {
		w.callback(*r.Value)
	}
}

func (w *refWriter) callback(c openapi.Callback) {
	for _, p := range c {
		w.pathItemRef(p)
	}
}

func (w *refWriter) parameterList(ps openapi.ParameterList) {
	for _, p := range ps {
		w.parameter(p)
	}
}

func (w *refWriter) parameter(r *openapi.ParameterRef) {
	if r == nil || r.Value == nil {
		return
	}

	w.schemaRef(r.Value.Schema)
	w.content(r.Value.Content)
}

func (w *refWriter) requestBody(r *openapi.RequestBodyRef) {
	if r == nil || r.Value == nil {
		return
	}

	w.content(r.Value.Content)
}

func (w *refWriter) response(r *openapi.ResponseRef) {
	if r == nil || r.Value == nil {
		return
	}

	w.headers(r.Value.Headers)
	w.content(r.Value.Content)
}

func (w *refWriter) headers(hs openapi.Headers) {
	for _, r := range hs {
		if r == nil || r.Value == nil {
			continue
		}

		w.schema(r.Value.Schema)
		w.content(r.Value.Content)
	}
}

func (w *refWriter) content(c openapi.Content) {
	for _, mt := range c {
		if mt == nil {
			continue
		}

		w.schemaRef(mt.Schema)

		for _, e := range mt.Encoding {
			if e != nil {
				w.headers(e.Headers)
			}
		}
	}
}

func (w *refWriter) schemaRef(r *openapi.SchemaRef) {
	if r == nil {
		return
	}

	if r.Ref != nil && r.Ref.Identifier == w.old {
		r.Ref.Identifier = w.new
	}

	// A resolved reference also carries the schema it points at. Walking it is
	// what reaches references nested inside a referenced schema.
	w.schema(r.Value)
}

func (w *refWriter) schemaRefList(l openapi.SchemaRefList) {
	for _, r := range l {
		w.schemaRef(r)
	}
}

func (w *refWriter) schema(s *openapi.Schema) {
	if s == nil || w.visited[s] {
		return
	}

	w.visited[s] = true

	w.schemaRefList(s.AllOf)
	w.schemaRefList(s.OneOf)
	w.schemaRefList(s.AnyOf)
	w.schemaRef(s.Not)
	w.schemaRef(s.Items)
	w.schemaRef(s.AdditionalProperties)

	for _, r := range s.Properties {
		w.schemaRef(r)
	}
}
