package edit

import "github.com/MarkRosemaker/openapi"

// MergeSchema repoints every reference to oldName at newName and removes
// oldName from components.schemas, leaving newName's own definition
// untouched.
//
// It differs from [RenameSchema], which refuses to rename a schema onto a
// name that already exists ([ErrSchemaExists]): merging onto an existing
// schema is exactly the point here, typically because several near-duplicate
// schemas (e.g. ones an OpenAPI generator produced one per endpoint, that
// happen to describe the same thing) are being consolidated into one.
//
// If description is non-empty, it becomes the $ref-level description on
// every reference this repoints, replacing whatever description that
// reference already had. The usual reason to reach for Merge rather than
// Rename is that oldName's own definition — its bounds, its wording — is
// about to be discarded once oldName is gone; setting description is how
// that information survives on the fields that used it, rather than being
// lost in the merge.
//
// It fails, changing nothing, if oldName or newName is not in
// components.schemas ([ErrSchemaNotFound]).
func MergeSchema(doc *openapi.Document, oldName, newName, description string) error {
	schemas := doc.Components.Schemas

	if _, ok := schemas[oldName]; !ok {
		return &ErrSchemaNotFound{Name: oldName}
	}

	if _, ok := schemas[newName]; !ok {
		return &ErrSchemaNotFound{Name: newName}
	}

	if oldName == newName {
		return nil
	}

	old, new := schemaRefPrefix+oldName, schemaRefPrefix+newName

	WalkSchemaRefs(doc, func(r *openapi.SchemaRef) {
		if r.Ref == nil || r.Ref.Identifier != old {
			return
		}

		if description != "" {
			r.Ref.Description = description
		}
		r.Ref.Identifier = new
	})

	delete(schemas, oldName)

	return nil
}
