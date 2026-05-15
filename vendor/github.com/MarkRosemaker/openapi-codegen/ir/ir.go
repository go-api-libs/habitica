package ir

import "github.com/MarkRosemaker/openapi"

// Document is the top-level IR type passed to templates.
type Document struct {
	PackageName       string
	BaseURL           URLParts
	UserAgent         string
	Operations        []Operation
	Schemas           []Schema
	Auth              Auth `json:"Security,omitzero"`
	HasURLFields      bool
	HasDurationFields bool
	HasDateFields     bool

	// InteractionCalls holds one entry per matched interaction.
	// Populated at code-gen time; not serialized to ir.json (too noisy).
	InteractionCalls []InteractionCall `json:"-"`
}

// InteractionCall is one operation call extracted from a recorded interaction.
type InteractionCall struct {
	Op        *Operation         // matched operation
	PathArgs  []string           // Go literal per path param, same order as Op.PathParams
	QueryArgs []InteractionParam // set query params only (omitted = use nil params)
}

// InteractionParam is one query param with its Go literal value.
type InteractionParam struct {
	FieldName string // PascalCase field name on the params struct
	Literal   string // Go expression, e.g. `3` or `"abc"`
}

// URLParts holds a decomposed server URL.
type URLParts struct {
	Scheme string
	Host   string
	Path   string
}

// Operation represents a single API operation.
type Operation struct {
	Name            string     `json:"Name,omitzero"`
	Description     string     `json:"Description,omitzero"`
	Summary         string     `json:"Summary,omitzero"`
	Method          string     `json:"Method,omitzero"`
	PathTemplate    string     `json:"PathTemplate,omitzero"`
	JoinPathArgs    []string   `json:"JoinPathArgs,omitempty"`
	PathParams      []Param    `json:"PathParams,omitempty"`
	QueryParams     []Param    `json:"QueryParams,omitempty"`
	HeaderParam     []Param    `json:"HeaderParam,omitempty"`
	HasParams       bool       `json:"HasParams,omitzero"`
	ParamStructName string     `json:"ParamStructName,omitzero"`
	RequestBody     *ReqBody   `json:"RequestBody,omitempty"`
	Responses       []Response `json:"Responses,omitempty"`
	SuccessReturn   *GoType    `json:"SuccessReturn,omitempty"`
	Deprecated      bool       `json:"Deprecated,omitzero"`
}

// Schema represents a named component schema.
type Schema struct {
	Name        string      `json:"Name,omitzero"`
	Description string      `json:"Description,omitzero"`
	Kind        SchemaKind  `json:"Kind"`
	Type        string      `json:"Type,omitzero"`
	Fields      []Field     `json:"Fields,omitzero,omitempty"`
	EnumValues  []EnumValue `json:"EnumValues,omitzero,omitempty"`
	MapKey      string      `json:"MapKey,omitzero"`
	MapValue    string      `json:"MapValue,omitzero"`
}

// SchemaKind categorizes a schema into struct, enum, or array alias.
type SchemaKind int

const (
	SchemaKindStruct     SchemaKind = iota // object with properties
	SchemaKindEnum                         // string with enum values
	SchemaKindArrayAlias                   // array type alias
	SchemaKindAllOf                        // allOf composition (struct with embedded types)
	SchemaKindMap
)

// Field is a named field within a struct schema.
type Field struct {
	Name        string
	JSONName    string
	Type        string
	JSONTag     string
	Description string
	Required    bool
	Embedded    bool `json:",omitzero"` // true for allOf $ref entries rendered as embedded structs
}

// EnumValue is one member of an enum type.
type EnumValue struct {
	GoName string
	Value  string
}

// Param represents a path or query parameter.
type Param struct {
	GoName       string // camelCase, used as local variable name
	FieldName    string // PascalCase, used as exported struct field name
	JSONName     string
	Type         string
	Required     bool
	NotZero      string
	FormatExpr   string
	ParseExpr    string // server-side: expression to parse string `s` into the param type
	ParseCast    string // server-side: optional cast after ParseExpr, e.g. "int32"
	ParseErrFree bool   // server-side: true when ParseExpr cannot return an error
	IsEnum       bool
	Description  string
	Value        string `json:"Value,omitzero"` // hardcoded value
}

// GoType is a resolved Go type reference.
type GoType struct {
	Name      string
	IsPointer bool
	IsSlice   bool
}

// String returns the Go type expression.
func (t GoType) String() string {
	switch {
	case t.IsPointer:
		return "*" + t.Name
	case t.IsSlice:
		return "[]" + t.Name
	default:
		return t.Name
	}
}

func (t GoType) Nilable() string {
	switch {
	case t.IsSlice:
		return "[]" + t.Name
	default:
		return "*" + t.Name
	}
}

// ZeroValue returns the Go zero-value literal for the type.
func (t GoType) ZeroValue() string {
	if t.IsPointer || t.IsSlice {
		return "nil"
	}
	switch t.Name {
	case "string":
		return `""`
	case "bool":
		return "false"
	case "int", "int32", "int64", "uint", "uint32", "uint64", "float32", "float64":
		return "0"
	case "uuid.UUID":
		return "uuid.Nil"
	default:
		return t.Name + "{}"
	}
}

// Response represents one expected HTTP response from an operation.
type Response struct {
	StatusCode  string
	GoConstant  string
	Description string
	ContentType string
	GoType      *GoType
	IsSuccess   bool
}

// ReqBody is the IR representation of an operation request body.
type ReqBody struct {
	TypeName    string
	ContentType string
	Required    bool
}

type Auth struct {
	Bearer Bearer `json:"Bearer,omitzero"`
	APIKey APIKey `json:"APIKey,omitzero"`
}

type Bearer struct {
	Name string
}

type APIKey struct {
	EnvName string
	Name    string
	In      openapi.ParameterLocation
	Example string
}
