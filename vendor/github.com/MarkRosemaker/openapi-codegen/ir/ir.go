package ir

import (
	"fmt"
	"slices"
	"strings"
)

// Document is the top-level IR type passed to templates.
type Document struct {
	Title                  string      `json:"title,omitzero"`
	Production             bool        `json:"production,omitzero"`
	PackageName            string      `json:"packageName,omitzero"`
	BaseURL                URLParts    `json:"baseURL,omitzero"`
	UserAgent              string      `json:"userAgent,omitzero"`
	Operations             []Operation `json:"operations,omitempty"`
	GlobalParams           Params      `json:"globalParams,omitempty"`
	Schemas                []Schema    `json:"schemas,omitempty"`
	Auth                   Auth        `json:"security,omitzero"`
	HasURLFields           bool        `json:"hasURLFields,omitzero"`
	HasDurationFields      bool        `json:"hasDurationFields,omitzero"`
	HasDateFields          bool        `json:"hasDateFields,omitzero"`
	HasDateTimeOrIntFields bool        `json:"hasDateTimeOrIntFields,omitzero"`

	// InteractionCalls holds one entry per matched interaction.
	// Populated at code-gen time; not serialized to ir.json (too noisy).
	InteractionCalls []InteractionCall `json:"-"`
}

// InteractionCall is one operation call extracted from a recorded interaction.
type InteractionCall struct {
	Op         *Operation         // matched operation
	PathArgs   []string           // Go literal per path param, same order as Op.PathParams
	QueryArgs  []InteractionParam // set query params only (omitted = use nil params)
	HeaderArgs []InteractionParam // set query params only (omitted = use nil params)
}

// InteractionParam is one query param with its Go literal value.
type InteractionParam struct {
	FieldName string // PascalCase field name on the params struct
	Literal   string // Go expression, e.g. `3` or `"abc"`
}

// URLParts holds a decomposed server URL.
type URLParts struct {
	Scheme string `json:"scheme,omitzero"`
	Host   string `json:"host,omitzero"`
	Path   string `json:"path,omitzero"`
}

// Operation represents a single API operation.
type Operation struct {
	Name            string     `json:"name,omitzero"`
	Description     string     `json:"description,omitzero"`
	Summary         string     `json:"summary,omitzero"`
	Method          string     `json:"method,omitzero"`
	PathTemplate    string     `json:"pathTemplate,omitzero"`
	JoinPathArgs    []string   `json:"joinPathArgs,omitempty"`
	PathParams      Params     `json:"pathParams,omitempty"`
	QueryParams     Params     `json:"queryParams,omitempty"`
	HeaderParams    Params     `json:"headerParams,omitempty"`
	HasParams       bool       `json:"hasParams,omitzero"`
	ParamStructName string     `json:"paramStructName,omitzero"`
	RequestBody     *ReqBody   `json:"requestBody,omitempty"`
	Responses       []Response `json:"responses,omitempty"`
	SuccessReturn   *GoType    `json:"successReturn,omitempty"`
	Deprecated      bool       `json:"deprecated,omitzero"`
}

func (op Operation) ParamsInStruct() Params {
	return append(op.QueryParams, op.HeaderParams...)
}

func (op Operation) NilParamsExpr() string {
	params := op.ParamsInStruct()
	if len(params) == 0 {
		return ""
	}

	if params.Required() {
		return fmt.Sprintf("%s{}", op.ParamStructName)
	}

	return "nil"
}

// JSPathTemplate returns the path template with {jsonName} placeholders replaced
// by ${goName} JavaScript template-literal interpolations.
func (op Operation) JSPathTemplate() string {
	result := op.PathTemplate
	for _, p := range op.PathParams {
		result = strings.ReplaceAll(result, "{"+p.JSONName+"}", "${"+p.GoName+"}")
	}
	return result
}

// Schema represents a named component schema.
type Schema struct {
	Name        string      `json:"name,omitzero"`
	Description string      `json:"description,omitzero"`
	Kind        SchemaKind  `json:"kind,omitzero"`
	Type        string      `json:"type,omitzero"`
	Fields      []Field     `json:"fields,omitempty"`
	EnumValues  []EnumValue `json:"enumValues,omitempty"`
	MapKey      string      `json:"mapKey,omitzero"`
	MapValue    string      `json:"mapValue,omitzero"`
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
	Name        string `json:"name,omitzero"`
	JSONName    string `json:"jsonName,omitzero"`
	Type        string `json:"type,omitzero"`
	JSONTag     string `json:"jsonTag,omitzero"`
	Description string `json:"description,omitzero"`
	Required    bool   `json:"required,omitzero"`
	Embedded    bool   `json:"embedded,omitzero"` // true for allOf $ref entries rendered as embedded structs

	// IsDateTimeOrInt is true when the property's schema is a oneOf of a
	// date-time string and an integer. The Go type is time.Time, but a custom
	// (un)marshaller is required to accept either form on the wire.
	IsDateTimeOrInt bool `json:"isDateTimeOrInt,omitzero"`
}

// EnumValue is one member of an enum type.
type EnumValue struct {
	GoName string `json:"goName,omitzero"`
	Value  string `json:"value,omitzero"`
}

type GlobalType string

const (
	GlobalAPIKey    GlobalType = "APIKey"
	GlobalVersion   GlobalType = "Version"
	GlobalClient    GlobalType = "Client"
	GlobalUserAgent GlobalType = "User-Agent"
)

type Params []Param

func (ps Params) Required() bool {
	for _, p := range ps {
		if p.Required {
			return true
		}
	}

	return false
}

// Param represents a path or query parameter.
type Param struct {
	GlobalType   GlobalType `json:"globalType,omitzero"`
	VarName      string     `json:"varName,omitzero"`
	EnvName      string     `json:"envName,omitzero"`
	GoName       string     `json:"goName,omitzero"`
	FieldName    string     `json:"fieldName,omitzero"`
	JSONName     string     `json:"jsonName,omitzero"`
	Type         string     `json:"type,omitzero"`
	Required     bool       `json:"required,omitzero"`
	ParseExpr    string     `json:"parseExpr,omitzero"`
	ParseCast    string     `json:"parseCast,omitzero"`
	ParseErrFree bool       `json:"parseErrFree,omitzero"`
	IsEnum       bool       `json:"isEnum,omitzero"`
	Description  string     `json:"description,omitzero"`
	Value        string     `json:"value,omitzero"`   // hardcoded value, always the same
	Example      string     `json:"example,omitzero"` // hardcoded example for tests
}

func (doc Document) APIKey() *Param {
	return doc.getGlobal(GlobalAPIKey)
}

func (doc Document) Client() *Param {
	return doc.getGlobal(GlobalClient)
}

func (doc Document) getGlobal(tp GlobalType) *Param {
	if i := slices.IndexFunc(doc.GlobalParams, func(p Param) bool {
		return p.GlobalType == tp
	}); i > -1 {
		p := doc.GlobalParams[i]
		return &p
	}

	return nil
}

// GoType is a resolved Go type reference.
type GoType struct {
	Name      string `json:"name,omitzero"`
	IsPointer bool   `json:"isPointer,omitzero"`
	IsSlice   bool   `json:"isSlice,omitzero"`
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
	StatusCode  string  `json:"statusCode,omitzero"`
	GoConstant  string  `json:"goConstant,omitzero"`
	Description string  `json:"description,omitzero"`
	ContentType string  `json:"contentType,omitzero"`
	GoType      *GoType `json:"goType,omitempty"`
	IsSuccess   bool    `json:"isSuccess,omitzero"`
}

// ReqBody is the IR representation of an operation request body.
type ReqBody struct {
	TypeName    string `json:"typeName,omitzero"`
	ContentType string `json:"contentType,omitzero"`
	Required    bool   `json:"required,omitzero"`
}

type Auth struct {
	Bearer Bearer `json:"bearer,omitzero"`
}

type Bearer struct {
	Name string `json:"name,omitzero"`
}
