package ir

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MarkRosemaker/openapi"
	"github.com/ettle/strcase"
)

// FromOperation converts an openapi operation to its IR representation.
// pathItemParams are the parameters defined at the path item level and are merged
// with (and can be overridden by) operation-level parameters.
func FromOperation(
	rawPath openapi.Path,
	pathItemParams openapi.ParameterList,
	method string,
	op *openapi.Operation,
	globalParams paramMap,
) (*Operation, error) {
	if op.OperationID == "" {
		return nil, fmt.Errorf("operationId is required")
	}

	name := strcase.ToGoPascal(op.OperationID)

	// Merge path-item parameters with operation parameters (operation overrides on conflict).
	merged := mergeParams(pathItemParams, op.Parameters)

	parsedPath := rawPath.Parse()

	// Resolve each parameter and index by name for path arg computation.
	var pathParams, queryParams, headerParams []Param
	paramByName := make(map[string]Param, len(merged))

	for _, pRef := range merged {
		p := pRef.Value
		if _, ok := globalParams[p]; ok {
			continue
		}

		param, err := fromParam(p, "")
		if err != nil {
			return nil, fmt.Errorf("param %q: %w", p.Name, err)
		}
		paramByName[p.Name] = param

		switch p.In {
		case openapi.ParameterLocationPath:
			pathParams = append(pathParams, param)
		case openapi.ParameterLocationQuery:
			queryParams = append(queryParams, param)
		case openapi.ParameterLocationHeader:
			headerParams = append(headerParams, param)
		}
	}

	joinArgs := buildJoinPathArgs(parsedPath, paramByName)

	hasParams := len(pathParams)+len(queryParams)+len(headerParams) > 0
	var paramStructName string
	if len(queryParams)+len(headerParams) > 0 {
		paramStructName = name + "Params"
	}

	var reqBody *ReqBody
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		var err error
		reqBody, err = fromRequestBody(op.RequestBody.Value)
		if err != nil {
			return nil, fmt.Errorf("requestBody: %w", err)
		}
	}

	responses, successReturn, err := fromResponses(op.Responses)
	if err != nil {
		return nil, fmt.Errorf("responses: %w", err)
	}

	return &Operation{
		Name:            name,
		Description:     op.Description,
		Summary:         op.Summary,
		Method:          strings.ToUpper(method),
		PathTemplate:    string(rawPath),
		JoinPathArgs:    joinArgs,
		PathParams:      pathParams,
		QueryParams:     queryParams,
		HeaderParams:    headerParams,
		HasParams:       hasParams,
		ParamStructName: paramStructName,
		RequestBody:     reqBody,
		Responses:       responses,
		SuccessReturn:   successReturn,
		Deprecated:      op.Deprecated,
	}, nil
}

// mergeParams merges path-item params with operation params; operation wins on (name, in) collision.
func mergeParams(pathItem, operation openapi.ParameterList) openapi.ParameterList {
	if len(pathItem) == 0 {
		return operation
	}

	overrides := make(map[string]bool, len(operation))
	for _, pRef := range operation {
		overrides[string(pRef.Value.In)+":"+pRef.Value.Name] = true
	}

	result := make(openapi.ParameterList, 0, len(pathItem)+len(operation))
	for _, pRef := range pathItem {
		if !overrides[string(pRef.Value.In)+":"+pRef.Value.Name] {
			result = append(result, pRef)
		}
	}
	return append(result, operation...)
}

func fromParam(p *openapi.Parameter, apiTitle string) (Param, error) {
	param := Param{
		JSONName:    p.Name,
		Required:    p.Required,
		Description: p.Description,
	}
	if p.Schema == nil {
		return param, fmt.Errorf("schema is required")
	}

	param.IsEnum = len(p.Schema.Enum) > 0

	tp, err := SchemaGoType(p.Schema)
	if err != nil {
		return param, err
	}
	param.Type = tp.String()

	param.GoName = strcase.ToGoCamel(p.Name)
	param.ParseExpr, param.ParseCast, param.ParseErrFree = tp.serverParseExpr()

	param.FieldName = strcase.ToGoPascal(p.Name)

	if p.Required {
		switch p.Name {
		case "X-Api-Key":
			param.GlobalType = GlobalAPIKey
			param.VarName = "apiKey"
			if apiTitle != "" {
				param.EnvName = strcase.ToSNAKE(fmt.Sprintf("%s_KEY", apiTitle))
			}
		case "X-Client":
			param.GlobalType = GlobalClient
			param.VarName = "client"
		default:
			if p.In == openapi.ParameterLocationHeader &&
				strings.HasSuffix(p.Name, "Version") {
				param.GlobalType = GlobalVersion
				param.Value = p.Schema.Example.String()
			}
		}
	}

	if param.VarName == "" {
		if p.In == openapi.ParameterLocationPath {
			param.VarName = strcase.ToGoCamel(p.Name)
		} else {
			param.VarName = "params." + param.FieldName
		}
	}

	if param.GlobalType != "" {
		param.Example = p.Schema.Example.String()
	}

	return param, nil
}

// serverParseExpr returns the expression that parses a string variable `s` into goType.
// cast is a non-empty type name when the parse result needs casting (e.g. int32 from ParseInt).
// errFree is true when the expression cannot return an error.
func (tp GoType) serverParseExpr() (expr, cast string, errFree bool) {
	switch tp.Name {
	case "string":
		return "s", "", true
	case "types.Email":
		return "types.Email(s)", "", true
	case "bool":
		return "strconv.ParseBool(s)", "", false
	case "int":
		return "strconv.Atoi(s)", "", false
	case "int32":
		return "strconv.ParseInt(s, 10, 32)", "int32", false
	case "int64":
		return "strconv.ParseInt(s, 10, 64)", "", false
	case "uint":
		return "strconv.ParseUint(s, 10, 64)", "uint", false
	case "uint32":
		return "strconv.ParseUint(s, 10, 32)", "uint32", false
	case "uint64":
		return "strconv.ParseUint(s, 10, 64)", "", false
	case "float32":
		return "strconv.ParseFloat(s, 32)", "float32", false
	case "float64":
		return "strconv.ParseFloat(s, 64)", "", false
	case "uuid.UUID":
		return "uuid.Parse(s)", "", false
	case "time.Time":
		return "time.Parse(time.RFC3339, s)", "", false
	case "civil.Date":
		return "civil.ParseDate(s)", "", false
	case "net.IP":
		return "net.ParseIP(s)", "", true
	case "time.Duration":
		return "time.ParseDuration(s)", "", false
	default:
		// string-based enum or other cast from string
		return tp.Name + "(s)", "", true
	}
}

// buildJoinPathArgs produces the ordered list of Go expressions for url.JoinPath.
// e.g. "/apis/{id}/items" → [`"apis"`, `strconv.Itoa(id)`, `"items"`]
func buildJoinPathArgs(parsed openapi.ParsedPath, params map[string]Param) []string {
	segments := strings.Split(strings.TrimPrefix(parsed.String(), "/"), "/")
	args := make([]string, 0, len(segments))
	for _, seg := range segments {
		if seg == "" {
			continue
		}

		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			paramName := seg[1 : len(seg)-1]
			if p, ok := params[paramName]; ok {
				args = append(args, p.FormatExpr())
			} else {
				args = append(args, strconv.Quote(paramName))
			}
		} else {
			args = append(args, strconv.Quote(seg))
		}
	}

	return args
}

// NotZero returns the Go boolean expression that is true when param is not the zero value.
func (p Param) NotZero() string {
	switch p.Type {
	case "string":
		return p.VarName + ` != ""`
	case "types.Email":
		return p.VarName + ` != ""`
	case "bool":
		return p.VarName
	case "uuid.UUID":
		return p.VarName + " != uuid.Nil"
	case "net.IP":
		return p.VarName + " != nil"
	case "url.URL":
		return p.VarName + `.Host != ""`
	case "time.Time":
		return "!" + p.VarName + ".IsZero()"
	case "civil.Date":
		return p.VarName + " != (civil.Date{})"
	case "time.Duration":
		return p.VarName + " != 0"
	default:
		// int, int32, int64, uint*, float32, float64
		return p.VarName + " != 0"
	}
}

// formatExpr returns the Go expression that converts the param to a string for URL encoding.
func (p Param) FormatExpr() string {
	if p.GlobalType != "" {
		return "c." + p.VarName
	}

	switch p.Type {
	case "string":
		return p.VarName
	case "types.Email":
		return "string(" + p.VarName + ")"
	case "bool":
		return "strconv.FormatBool(" + p.VarName + ")"
	case "int":
		return "strconv.Itoa(" + p.VarName + ")"
	case "int32":
		return "strconv.FormatInt(int64(" + p.VarName + "), 10)"
	case "int64":
		return "strconv.FormatInt(" + p.VarName + ", 10)"
	case "uint":
		return "strconv.FormatUint(uint64(" + p.VarName + "), 10)"
	case "uint32":
		return "strconv.FormatUint(uint64(" + p.VarName + "), 10)"
	case "uint64":
		return "strconv.FormatUint(" + p.VarName + ", 10)"
	case "float32":
		return "strconv.FormatFloat(float64(" + p.VarName + "), 'f', -1, 32)"
	case "float64":
		return "strconv.FormatFloat(" + p.VarName + ", 'f', -1, 64)"
	case "uuid.UUID":
		return p.VarName + ".String()"
	case "url.URL":
		return p.VarName + ".String()"
	case "time.Time":
		return p.VarName + ".Format(time.RFC3339)"
	case "civil.Date":
		return p.VarName + ".String()"
	case "net.IP":
		return p.VarName + ".String()"
	case "time.Duration":
		return "strconv.FormatInt(int64(" + p.VarName + "/time.Second), 10)"
	default:
		return "fmt.Sprint(" + p.VarName + ")"
	}
}

func fromRequestBody(rb *openapi.RequestBody) (*ReqBody, error) {
	for mr, mt := range rb.Content.ByIndex() {
		if mt.Schema == nil {
			continue
		}
		tp, err := SchemaRefGoType(mt.Schema)
		if err != nil {
			return nil, err
		}
		return &ReqBody{
			TypeName:    tp.String(),
			ContentType: string(mr),
			Required:    rb.Required,
		}, nil
	}
	return nil, nil
}

func fromResponses(responses openapi.OperationResponses) ([]Response, *GoType, error) {
	var result []Response
	var successReturn *GoType

	for code, rRef := range responses.ByIndex() {
		r := rRef.Value

		isSuccess := code.IsSuccess()
		goConst := statusCodeToConst(code)

		var goType *GoType
		var contentType string

		for mr, mt := range r.Content.ByIndex() {
			if !strings.Contains(string(mr), "json") {
				continue
			}
			contentType = string(mr)
			if mt.Schema != nil {
				var err error
				goType, err = SchemaRefGoType(mt.Schema)
				if err != nil {
					return nil, nil, fmt.Errorf("response %s: %w", code, err)
				}
			}
			break
		}

		result = append(result, Response{
			StatusCode:  string(code),
			GoConstant:  goConst,
			Description: r.Description,
			ContentType: contentType,
			GoType:      goType,
			IsSuccess:   isSuccess,
		})

		if isSuccess && goType != nil && successReturn == nil {
			successReturn = goType
		}
	}

	return result, successReturn, nil
}

// statusCodeToConst converts an OpenAPI status code to its net/http constant name.
func statusCodeToConst(code openapi.StatusCode) string {
	if code == openapi.StatusCodeDefault {
		return "0"
	}
	n, err := strconv.Atoi(string(code))
	if err != nil {
		return string(code)
	}
	text := http.StatusText(n)
	if text == "" {
		return string(code)
	}
	// "No Content" → "NoContent" → "http.StatusNoContent"
	return "http.Status" + strings.ReplaceAll(text, " ", "")
}
