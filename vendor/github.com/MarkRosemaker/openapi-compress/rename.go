package compress

import (
	"fmt"

	"github.com/MarkRosemaker/openapi"
)

// RenameSchema renames a schema in the document and updates every $ref
// identifier throughout the document that points to the old name.
func RenameSchema(d *openapi.Document, oldName, newName string) error {
	if oldName == newName {
		return nil
	}
	s, ok := d.Components.Schemas[oldName]
	if !ok {
		return fmt.Errorf("schema %q not found in components.schemas", oldName)
	}
	if _, exists := d.Components.Schemas[newName]; exists {
		return fmt.Errorf("schema %q already exists in components.schemas", newName)
	}
	delete(d.Components.Schemas, oldName)
	d.Components.Schemas.Set(newName, s)
	replaceRefsInDocument(d, map[string]string{oldName: newName})
	return nil
}
