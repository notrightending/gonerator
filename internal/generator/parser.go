// Package generator provides functionality to parse Go source files and generate API handlers.
package generator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// ApiMethod represents the API method configuration extracted from comments.
type ApiMethod struct {
	Url        string `json:"url"`
	Auth       bool   `json:"auth"`
	Method     string `json:"method"`
	AuthEnvKey string `json:"auth_env_key"`
}

// ApiValidatorTag represents the validation rules for API parameters.
type ApiValidatorTag struct {
	Required  bool
	Min       *int
	Max       *int
	ParamName string
	Enum      []string
	Default   string
}

// StructField represents a field in the input struct for an API method.
type StructField struct {
	Name string
	Type string
	Tag  ApiValidatorTag
}

// Method represents a parsed API method with all its metadata.
type Method struct {
	Name         string
	ReceiverName string
	ReceiverType string
	InputType    string
	OutputType   string
	ApiMethod    ApiMethod
	StructFields []StructField
}

// parseFile parses the given Go source file and extracts API method information.
func parseFile(filename string) ([]Method, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var methods []Method

	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Doc != nil {
				for _, comment := range funcDecl.Doc.List {
					if strings.HasPrefix(comment.Text, "// apigen:api") {
						method, err := parseMethod(funcDecl, comment.Text, filename)
						if err != nil {
							return nil, err
						}
						methods = append(methods, method)
						break
					}
				}
			}
		}
	}

	return methods, nil
}

// parseMethod extracts method information from an AST function declaration.
func parseMethod(funcDecl *ast.FuncDecl, comment, filename string) (Method, error) {
	method := Method{
		Name:         funcDecl.Name.Name,
		ReceiverName: funcDecl.Recv.List[0].Names[0].Name,
		ReceiverType: funcDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name,
		InputType:    funcDecl.Type.Params.List[1].Type.(*ast.Ident).Name,
		OutputType:   funcDecl.Type.Results.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name,
	}

	apiMethod := ApiMethod{}
	err := json.Unmarshal([]byte(strings.TrimPrefix(comment, "// apigen:api")), &apiMethod)
	if err != nil {
		return Method{}, err
	}
	method.ApiMethod = apiMethod

	// Set default method to GET,POST if not specified
	if method.ApiMethod.Method == "" {
		method.ApiMethod.Method = "GET,POST"
	}

	// Set default auth env key if auth is required but no key is specified
	if method.ApiMethod.Auth && method.ApiMethod.AuthEnvKey == "" {
		method.ApiMethod.AuthEnvKey = "API_AUTH_KEY"
	}

	structFields, err := parseStructFields(filename, method.InputType)
	if err != nil {
		return Method{}, err
	}
	method.StructFields = structFields

	return method, nil
}

// parseStructFields extracts field information from the input struct of an API method.
func parseStructFields(filename string, structName string) ([]StructField, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var fields []StructField

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == structName {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					for _, field := range structType.Fields.List {
						if len(field.Names) > 0 {
							fieldName := field.Names[0].Name
							fieldType := fmt.Sprintf("%s", field.Type)
							tag := parseApiValidatorTag(field.Tag)
							fields = append(fields, StructField{
								Name: fieldName,
								Type: fieldType,
								Tag:  tag,
							})
						}
					}
				}
			}
		}
		return true
	})

	return fields, nil
}

// parseApiValidatorTag parses the apivalidator tag and extracts validation rules.
func parseApiValidatorTag(tag *ast.BasicLit) ApiValidatorTag {
	if tag == nil {
		return ApiValidatorTag{}
	}

	tagValue := strings.Trim(tag.Value, "`")
	apiValidatorTag := strings.TrimPrefix(tagValue, "apivalidator:")
	apiValidatorTag = strings.Trim(apiValidatorTag, "\"")

	parts := strings.Split(apiValidatorTag, ",")
	result := ApiValidatorTag{}

	for _, part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		key := keyValue[0]
		var value string
		if len(keyValue) > 1 {
			value = keyValue[1]
		}

		switch key {
		case "required":
			result.Required = true
		case "paramname":
			result.ParamName = value
		case "enum":
			result.Enum = strings.Split(value, "|")
		case "default":
			result.Default = value
		case "min":
			if intValue, err := strToInt(value); err == nil {
				result.Min = &intValue
			}
		case "max":
			if intValue, err := strToInt(value); err == nil {
				result.Max = &intValue
			}
		}
	}

	return result
}

func strToInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
