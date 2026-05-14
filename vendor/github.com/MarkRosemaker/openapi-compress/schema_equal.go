package compress

import (
	"reflect"
	"slices"

	"github.com/MarkRosemaker/openapi"
)

func schemasEqual(a, b *openapi.Schema) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return a.Title == b.Title &&
		a.Description == b.Description &&
		a.Type == b.Type &&
		a.Format == b.Format &&
		schemaRefListsEqual(a.AllOf, b.AllOf) &&
		float64PtrsEqual(a.Min, b.Min) &&
		float64PtrsEqual(a.Max, b.Max) &&
		regexpStringsEqual(a.Pattern, b.Pattern) &&
		slices.Equal(a.Enum, b.Enum) &&
		a.MinItems == b.MinItems &&
		uintPtrsEqual(a.MaxItems, b.MaxItems) &&
		schemaRefPtrsEqual(a.Items, b.Items) &&
		schemaRefsEqual(a.Properties, b.Properties) &&
		slices.Equal(a.Required, b.Required) &&
		schemaRefPtrsEqual(a.AdditionalProperties, b.AdditionalProperties) &&
		a.ContentMediaType == b.ContentMediaType &&
		a.ContentEncoding == b.ContentEncoding &&
		reflect.DeepEqual(a.Default, b.Default) &&
		// Example is documentation only; it does not affect schema validity.
		reflect.DeepEqual(a.Extensions, b.Extensions)
}

func schemaRefPtrsEqual(a, b *openapi.SchemaRef) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return schemaRefEqual(a, b)
}

func schemaRefEqual(a, b *openapi.SchemaRef) bool {
	switch {
	case a.Ref != nil && b.Ref != nil:
		return a.Ref.Identifier == b.Ref.Identifier &&
			a.Ref.Summary == b.Ref.Summary &&
			a.Ref.Description == b.Ref.Description
	case a.Ref == nil && b.Ref == nil:
		return schemasEqual(a.Value, b.Value)
	default:
		return false
	}
}

func schemaRefListsEqual(a, b openapi.SchemaRefList) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !schemaRefEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func schemaRefsEqual(a, b openapi.SchemaRefs) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if !schemaRefEqual(va, vb) {
			return false
		}
	}
	return true
}

func float64PtrsEqual(a, b *float64) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func uintPtrsEqual(a, b *uint) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func regexpStringsEqual(a, b interface{ String() string }) bool {
	if a == b {
		return true
	}
	if reflect.ValueOf(a).IsNil() != reflect.ValueOf(b).IsNil() {
		return false
	}
	if reflect.ValueOf(a).IsNil() {
		return true
	}
	return a.String() == b.String()
}
