package ir

import (
	"cmp"
	"fmt"
	"strings"
	"unicode"

	"github.com/MarkRosemaker/openapi"
	"github.com/ettle/strcase"
)

// SchemaRefGoType returns the Go type string for a SchemaRef.
// After flattening, complex schemas are moved to components and referenced by $ref;
// this function extracts the component name from the identifier or maps the inline type.
func SchemaRefGoType(ref *openapi.SchemaRef) (*GoType, error) {
	if ref.Ref != nil {
		// "#/components/schemas/Name" → "Name"
		parts := strings.Split(ref.Ref.Identifier, "/")
		return &GoType{Name: parts[len(parts)-1]}, nil
	}

	return SchemaGoType(ref.Value)
}

// SchemaGoType maps an openapi.Schema to its Go type string.
func SchemaGoType(s *openapi.Schema) (*GoType, error) {
	switch s.Type {
	case openapi.TypeBoolean:
		return &GoType{Name: "bool"}, nil
	case openapi.TypeInteger:
		return integerGoType(s.Format)
	case openapi.TypeNumber:
		return numberGoType(s.Format)
	case openapi.TypeString:
		return stringGoType(s.Format)
	case openapi.TypeArray:
		return arrayGoType(s)
	case openapi.TypeObject:
		return objectGoType(s)
	default:
		return nil, fmt.Errorf("unsupported schema type: %q", s.Type)
	}
}

func integerGoType(f openapi.Format) (*GoType, error) {
	switch f {
	case "":
		return &GoType{Name: "int"}, nil
	case openapi.FormatInt32:
		return &GoType{Name: "int32"}, nil
	case openapi.FormatInt64:
		return &GoType{Name: "int64"}, nil
	case openapi.FormatUint:
		return &GoType{Name: "uint"}, nil
	case openapi.FormatUint32:
		return &GoType{Name: "uint32"}, nil
	case openapi.FormatUint64:
		return &GoType{Name: "uint64"}, nil
	case openapi.FormatDuration:
		return &GoType{Name: "time.Duration"}, nil
	default:
		return nil, fmt.Errorf("unsupported integer format: %q", f)
	}
}

func numberGoType(f openapi.Format) (*GoType, error) {
	switch f {
	case "", openapi.FormatDouble:
		return &GoType{Name: "float64"}, nil
	case openapi.FormatFloat:
		return &GoType{Name: "float32"}, nil
	default:
		return nil, fmt.Errorf("unsupported number format: %q", f)
	}
}

func stringGoType(f openapi.Format) (*GoType, error) {
	switch f {
	case "", openapi.FormatPassword, openapi.FormatByte, openapi.FormatBinary, openapi.FormatZipCode:
		return &GoType{Name: "string"}, nil
	case openapi.FormatUUID:
		return &GoType{Name: "uuid.UUID"}, nil
	case openapi.FormatURI, openapi.FormatURIRef:
		return &GoType{Name: "url.URL"}, nil
	case openapi.FormatEmail:
		return &GoType{Name: "types.Email"}, nil
	case openapi.FormatDateTime:
		return &GoType{Name: "time.Time"}, nil
	case openapi.FormatDate:
		return &GoType{Name: "civil.Date"}, nil
	case openapi.FormatIPv4, openapi.FormatIPv6:
		return &GoType{Name: "net.IP"}, nil
	default:
		return nil, fmt.Errorf("unsupported string format: %q", f)
	}
}

func arrayGoType(s *openapi.Schema) (*GoType, error) {
	if s.Items == nil {
		return &GoType{Name: "[]any"}, nil
	}

	itemType, err := SchemaRefGoType(s.Items)
	if err != nil {
		return nil, fmt.Errorf("items: %w", err)
	}

	itemType.Name = "[]" + itemType.Name

	return itemType, nil
}

func objectGoType(s *openapi.Schema) (*GoType, error) {
	if s.AdditionalProperties != nil {
		tp, err := SchemaRefGoType(s.AdditionalProperties)
		if err != nil {
			return nil, fmt.Errorf("additionalProperties: %w", err)
		}

		return &GoType{Name: "map[string]" + tp.Name}, nil
	}

	// Named objects with properties are moved to components by the flatten pass.
	return &GoType{Name: "struct{}"}, nil
}

// FromComponentSchemas converts a set of named component schemas to IR schemas.
func FromComponentSchemas(schemas openapi.Schemas) ([]Schema, error) {
	result := make([]Schema, 0, len(schemas))
	for name, s := range schemas.ByIndex() {
		irSchema, err := fromSchema(name, s)
		if err != nil {
			return nil, fmt.Errorf("schema %q: %w", name, err)
		}

		if irSchema != nil {
			result = append(result, *irSchema)
		}
	}
	return result, nil
}

func fromSchema(name string, s *openapi.Schema) (*Schema, error) {
	switch s.Type {
	case openapi.TypeObject:
		if s.AdditionalProperties != nil {
			mapValueType, err := SchemaRefGoType(s.AdditionalProperties)
			if err != nil {
				return nil, err
			}

			if mapValueType.Name == "string" {
				return nil, nil
			}

			return &Schema{
				Name:        name,
				Description: getDescription(s, name),
				Kind:        SchemaKindMap,
				MapKey:      "string",
				MapValue:    mapValueType.String(),
			}, nil
		}

		return fromObjectSchema(name, s)
	case openapi.TypeString:
		if len(s.Enum) > 0 {
			return fromEnumSchema(name, s)
		}
		return nil, nil // plain strings don't produce named schemas
	case openapi.TypeArray:
		return fromArraySchema(name, s)
	case "":
		if len(s.AllOf) > 0 {
			return fromAllOfSchema(name, s)
		}
		return nil, nil
	default:
		return nil, nil // scalar types are used inline
	}
}

func getDescription(s *openapi.Schema, name string) string {
	if s.Description != "" {
		return s.Description
	}

	return fmt.Sprintf("%s defines a model", name)
}

func fromAllOfSchema(name string, s *openapi.Schema) (*Schema, error) {
	requiredSet := make(map[string]bool)
	for _, r := range s.Required {
		requiredSet[r] = true
	}

	var fields []Field
	for _, entry := range s.AllOf {
		if entry.Ref != nil {
			// $ref entry → embedded struct
			typeName, err := SchemaRefGoType(entry)
			if err != nil {
				return nil, err
			}

			fields = append(fields, Field{
				Type:     typeName.String(),
				Embedded: true,
			})
			continue
		}
		if entry.Value == nil {
			continue
		}

		// inline object entry → merge its properties as regular fields
		for _, r := range entry.Value.Required {
			requiredSet[r] = true
		}

		for jsonName, propRef := range entry.Value.Properties.ByIndex() {
			field, err := getField(jsonName, propRef, requiredSet)
			if err != nil {
				return nil, fmt.Errorf("allOf property %q: %w", jsonName, err)
			}

			fields = append(fields, field)
		}
	}

	return &Schema{
		Name:        name,
		Description: getDescription(s, name),
		Kind:        SchemaKindAllOf,
		Fields:      fields,
	}, nil
}

func getField(jsonName string, propRef *openapi.SchemaRef, requiredSet map[string]bool) (Field, error) {
	goType, err := SchemaRefGoType(propRef)
	if err != nil {
		return Field{}, err
	}

	v := propRef.Value

	required := requiredSet[jsonName]
	if !required {
		switch v.Type {
		case openapi.TypeBoolean, openapi.TypeArray:
		case openapi.TypeString:
			switch v.Format {
			case openapi.FormatURI, openapi.FormatUUID:
				goType.IsPointer = true
			}
		default:
			goType.IsPointer = true
		}
	}

	ref := cmp.Or(propRef.Ref, &openapi.Reference{})
	return Field{
		Name:        fieldGoName(jsonName),
		JSONName:    jsonName,
		Type:        goType.String(),
		JSONTag:     buildJSONTag(jsonName, v.Type, v.Format, required),
		Description: cmp.Or(ref.Description, v.Description),
		Required:    required,
	}, nil
}

func fromObjectSchema(name string, s *openapi.Schema) (*Schema, error) {
	requiredSet := make(map[string]bool, len(s.Required))
	for _, r := range s.Required {
		requiredSet[r] = true
	}

	fields := make([]Field, 0, len(s.Properties))
	for jsonName, propRef := range s.Properties.ByIndex() {
		field, err := getField(jsonName, propRef, requiredSet)
		if err != nil {
			return nil, fmt.Errorf("property %q: %w", jsonName, err)
		}

		fields = append(fields, field)
	}

	return &Schema{
		Name:        name,
		Description: getDescription(s, name),
		Kind:        SchemaKindStruct,
		Fields:      fields,
	}, nil
}

func fromEnumSchema(name string, s *openapi.Schema) (*Schema, error) {
	tp, err := stringGoType(s.Format)
	if err != nil {
		return nil, err
	}

	values := make([]EnumValue, len(s.Enum))
	for i, v := range s.Enum {
		values[i] = EnumValue{
			GoName: enumConstName(name, v),
			Value:  v,
		}
	}

	return &Schema{
		Name:        name,
		Description: getDescription(s, name),
		Kind:        SchemaKindEnum,
		Type:        tp.String(),
		EnumValues:  values,
	}, nil
}

func fromArraySchema(name string, s *openapi.Schema) (*Schema, error) {
	aliasType, err := arrayGoType(s)
	if err != nil {
		return nil, err
	}

	return &Schema{
		Name:        name,
		Description: getDescription(s, name),
		Kind:        SchemaKindArrayAlias,
		Type:        aliasType.String(),
	}, nil
}

// fieldGoName converts a JSON property name to an exported Go identifier.
func fieldGoName(jsonName string) string {
	// special case
	if strings.ToLower(jsonName) == "pdf" {
		return "PDF"
	}

	// replace special characters before PascalCasing
	r := strings.NewReplacer(
		"+", " Plus ",
		".", " Dot ",
		"/", " ",
		"(", "",
		")", "",
		"C#", "CSharp",
		"F#", "CSharp",
	)

	sanitized := r.Replace(jsonName)
	sanitized = replaceLeadingDigits(sanitized)
	return strcase.ToGoPascal(sanitized)
}

// enumConstName builds the Go constant name for an enum value, e.g. MyEnum + "foo_bar" → MyEnumFooBar.
func enumConstName(typeName, value string) string {
	r := strings.NewReplacer(
		"#", " Sharp ",
		"/", " ",
		"+", " Plus ",
		".", " Dot ",
		"(", "",
		")", "",
	)
	sanitized := r.Replace(value)
	sanitized = replaceLeadingDigits(sanitized)

	if len(sanitized) <= 3 && sanitized == strings.ToUpper(sanitized) {
		return typeName + sanitized
	}

	return typeName + strcase.ToGoPascal(sanitized)
}

// replaceLeadingDigits converts every leading digit in the first word to its
// word equivalent, so the result can be used as a Go identifier.
// e.g. "4K" → "Four K", "1080p" → "One Zero Eight Zero p".
func replaceLeadingDigits(s string) string {
	if s == "" {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	first := []rune(words[0])
	if len(first) == 0 || !unicode.IsDigit(first[0]) {
		return s
	}
	var parts []string
	i := 0
	for i < len(first) && unicode.IsDigit(first[i]) {
		parts = append(parts, digitWord(first[i]))
		i++
	}
	parts = append(parts, string(first[i:]))
	words[0] = strings.Join(parts, " ")
	return strings.Join(words, " ")
}

var digitWords = [10]string{
	"Zero", "One", "Two", "Three", "Four",
	"Five", "Six", "Seven", "Eight", "Nine",
}

func digitWord(r rune) string {
	d := int(r - '0')
	if d >= 0 && d < len(digitWords) {
		return digitWords[d]
	}
	return string(r)
}

// buildJSONTag computes the json struct tag for a field.
//
// Rules (mirroring the apilib reference):
//   - plain string, required:     json:"name"
//   - plain string, optional:     json:"name"       (empty strings are valid values)
//   - array:                      json:"name,omitempty"
//   - other, required:            json:"name,omitzero"
//   - other, optional:            json:"name,omitempty"
func buildJSONTag(jsonName string, tp openapi.DataType, format openapi.Format, required bool) string {
	// NOTE: JSON tags need to be rethought;
	// ideally, we want to not marshal unnecessarily
	// at the same time, sometimes we need to marshal null to delete values
	// we may need to decide based on custom x- tags in the openapi spec
	var opts string
	switch tp {
	case openapi.TypeString:
		switch format {
		case "": // regular string
			opts = ",omitzero"
		default:
			// NOTE: copied from legacy code, might not make sense
			if required {
				opts = ",omitzero"
			} else {
				opts = ",omitempty"
			}
		}
	case openapi.TypeArray:
		opts = ",omitempty" // always omit empty array
	case openapi.TypeBoolean, openapi.TypeInteger:
		// NOTE: copied from legacy code, might not make sense
		if required {
			opts = ",omitzero"
		} else {
			opts = ",omitempty"
		}
	default:
		if !required {
			opts = ",omitempty"
		}
	}

	return fmt.Sprintf(`json:"%s%s"`, jsonName, opts)
}
