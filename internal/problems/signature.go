package problems

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

var supportedValueTypes = map[string]bool{
	"string":         true,
	"bool":           true,
	"int":            true,
	"uint64":         true,
	"[]int":          true,
	"[]string":       true,
	"map[string]int": true,
}

var supportedCallbackTypes = map[string]bool{
	"func(int) int":                   true,
	"func(string) string":             true,
	"func(context.Context, int) bool": true,
}

type Parameter struct {
	Name string
	Type string
}

type SignatureSchema struct {
	Function string
	Params   []Parameter
	Result   string
}

func ParseSignature(signature string) (SignatureSchema, error) {
	source := "package solution\n" + strings.TrimSpace(signature) + " { panic(\"stub\") }\n"
	file, err := parser.ParseFile(token.NewFileSet(), "signature.go", source, parser.AllErrors)
	if err != nil {
		return SignatureSchema{}, fmt.Errorf("invalid function signature: %w", err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if function.Recv != nil || function.Name.Name != "Solve" {
			return SignatureSchema{}, errors.New("only a top-level Solve function is supported")
		}
		params, err := parseParameters(function.Type.Params)
		if err != nil {
			return SignatureSchema{}, err
		}
		results, err := flattenFieldTypes(function.Type.Results)
		if err != nil {
			return SignatureSchema{}, err
		}
		if len(results) != 1 {
			return SignatureSchema{}, errors.New("Solve must return exactly one value")
		}
		if len(params) == 0 {
			return SignatureSchema{}, errors.New("Solve must accept at least one argument")
		}
		for _, parameter := range params {
			if !supportedValueTypes[parameter.Type] && !supportedCallbackTypes[parameter.Type] {
				return SignatureSchema{}, fmt.Errorf("unsupported parameter type %s", parameter.Type)
			}
		}
		if !supportedValueTypes[results[0]] {
			return SignatureSchema{}, fmt.Errorf("unsupported result type %s", results[0])
		}
		return SignatureSchema{Function: "Solve", Params: params, Result: results[0]}, nil
	}
	return SignatureSchema{}, errors.New("required function Solve was not found")
}

func ValidateSolution(source, expectedSignature string) error {
	expected, err := ParseSignature(expectedSignature)
	if err != nil {
		return err
	}
	file, err := parser.ParseFile(token.NewFileSet(), "solution.go", source, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse Go source: %w", err)
	}
	if file.Name.Name != "solution" {
		return errors.New("package must be solution")
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name.Name != expected.Function {
			continue
		}
		actualParams, err := flattenFieldTypes(function.Type.Params)
		if err != nil {
			return fmt.Errorf("Solve signature must be %s", expectedSignature)
		}
		actualResults, err := flattenFieldTypes(function.Type.Results)
		if err != nil || len(actualResults) != 1 || actualResults[0] != expected.Result || len(actualParams) != len(expected.Params) {
			return fmt.Errorf("Solve signature must be %s", expectedSignature)
		}
		for index, parameter := range expected.Params {
			if actualParams[index] != parameter.Type {
				return fmt.Errorf("Solve signature must be %s", expectedSignature)
			}
		}
		return nil
	}
	return fmt.Errorf("required function %s was not found", expected.Function)
}

func ValidatePublicTests(schema SignatureSchema, tests []PublicTest) error {
	for testIndex, test := range tests {
		if len(test.Arguments) != len(schema.Params) {
			return fmt.Errorf("public test %d: got %d arguments, want %d", testIndex+1, len(test.Arguments), len(schema.Params))
		}
		for argumentIndex, argument := range test.Arguments {
			if _, err := GoLiteral(schema.Params[argumentIndex].Type, argument); err != nil {
				return fmt.Errorf("public test %d argument %s: %w", testIndex+1, schema.Params[argumentIndex].Name, err)
			}
		}
		if _, err := GoLiteral(schema.Result, test.Expected); err != nil {
			return fmt.Errorf("public test %d expected value: %w", testIndex+1, err)
		}
	}
	return nil
}

func GoLiteral(valueType string, raw json.RawMessage) (string, error) {
	switch valueType {
	case "string":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", errors.New("must be a string")
		}
		return strconv.Quote(value), nil
	case "bool":
		var value bool
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", errors.New("must be a boolean")
		}
		return strconv.FormatBool(value), nil
	case "int":
		var value int
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", errors.New("must be an integer")
		}
		return strconv.Itoa(value), nil
	case "uint64":
		var value uint64
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", errors.New("must be an unsigned integer")
		}
		return fmt.Sprintf("uint64(%d)", value), nil
	case "[]int":
		var values []int
		if err := json.Unmarshal(raw, &values); err != nil {
			return "", errors.New("must be an integer array")
		}
		parts := make([]string, len(values))
		for index, value := range values {
			parts[index] = strconv.Itoa(value)
		}
		return "[]int{" + strings.Join(parts, ", ") + "}", nil
	case "[]string":
		var values []string
		if err := json.Unmarshal(raw, &values); err != nil {
			return "", errors.New("must be a string array")
		}
		parts := make([]string, len(values))
		for index, value := range values {
			parts[index] = strconv.Quote(value)
		}
		return "[]string{" + strings.Join(parts, ", ") + "}", nil
	case "map[string]int":
		var values map[string]int
		if err := json.Unmarshal(raw, &values); err != nil {
			return "", errors.New("must be an object with integer values")
		}
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, strconv.Quote(key)+": "+strconv.Itoa(values[key]))
		}
		return "map[string]int{" + strings.Join(parts, ", ") + "}", nil
	case "func(int) int":
		var callback string
		if err := json.Unmarshal(raw, &callback); err != nil {
			return "", errors.New("must name a callback")
		}
		switch callback {
		case "identity":
			return "func(value int) int { return value }", nil
		case "square":
			return "func(value int) int { return value * value }", nil
		default:
			return "", fmt.Errorf("unsupported int callback %q", callback)
		}
	case "func(string) string":
		var callback string
		if err := json.Unmarshal(raw, &callback); err != nil || callback != "identity" {
			return "", errors.New("unsupported string callback")
		}
		return "func(value string) string { return value }", nil
	case "func(context.Context, int) bool":
		var callback string
		if err := json.Unmarshal(raw, &callback); err != nil || callback != "delay" {
			return "", errors.New("unsupported cancellable callback")
		}
		return `func(ctx context.Context, delay int) bool {
			timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
			defer timer.Stop()
			select {
			case <-timer.C:
				return true
			case <-ctx.Done():
				return false
			}
		}`, nil
	default:
		return "", fmt.Errorf("unsupported type %s", valueType)
	}
}

func DisplayJSON(raw json.RawMessage) string {
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return string(raw)
	}
	return compact.String()
}

func DisplayArguments(schema SignatureSchema, test PublicTest) string {
	parts := make([]string, 0, len(test.Arguments))
	for index, argument := range test.Arguments {
		name := fmt.Sprintf("arg%d", index+1)
		if index < len(schema.Params) && schema.Params[index].Name != "" {
			name = schema.Params[index].Name
		}
		parts = append(parts, name+" = "+DisplayJSON(argument))
	}
	return strings.Join(parts, "\n")
}

func publicValidationTestSource(schema SignatureSchema, tests []PublicTest) (string, error) {
	var cases strings.Builder
	for index, test := range tests {
		arguments := make([]string, len(test.Arguments))
		for argumentIndex, argument := range test.Arguments {
			literal, err := GoLiteral(schema.Params[argumentIndex].Type, argument)
			if err != nil {
				return "", err
			}
			arguments[argumentIndex] = literal
		}
		expected, err := GoLiteral(schema.Result, test.Expected)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(
			&cases,
			"\tif got, want := Solve(%s), %s; !reflect.DeepEqual(got, want) { t.Fatalf(\"public case %d failed: got %%v, want %%v\", got, want) }\n",
			strings.Join(arguments, ", "),
			expected,
			index+1,
		)
	}
	imports := "\t\"reflect\"\n\t\"testing\"\n"
	if strings.Contains(cases.String(), "context.") {
		imports = "\t\"context\"\n" + imports
	}
	if strings.Contains(cases.String(), "time.") {
		imports += "\t\"time\"\n"
	}
	return `package solution

import (
` + imports + `)

func TestCatalogPublic(t *testing.T) {
` + cases.String() + "}\n", nil
}

func parseParameters(fields *ast.FieldList) ([]Parameter, error) {
	if fields == nil {
		return nil, nil
	}
	params := make([]Parameter, 0, fields.NumFields())
	position := 0
	for _, field := range fields.List {
		valueType, err := expressionString(field.Type)
		if err != nil {
			return nil, err
		}
		if len(field.Names) == 0 {
			position++
			params = append(params, Parameter{Name: fmt.Sprintf("arg%d", position), Type: valueType})
			continue
		}
		for _, name := range field.Names {
			position++
			params = append(params, Parameter{Name: name.Name, Type: valueType})
		}
	}
	return params, nil
}

func flattenFieldTypes(fields *ast.FieldList) ([]string, error) {
	if fields == nil {
		return nil, nil
	}
	result := make([]string, 0, fields.NumFields())
	for _, field := range fields.List {
		valueType, err := expressionString(field.Type)
		if err != nil {
			return nil, err
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			result = append(result, valueType)
		}
	}
	return result, nil
}

func expressionString(expression ast.Expr) (string, error) {
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), expression); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
