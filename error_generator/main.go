package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/leshless/golibrary/stringcase"
)

//go:embed error.go.tpl
var errorTemplate string
var tmpl = template.Must(template.New("error").Parse(errorTemplate))

type ErrorField struct {
	Name        string
	Type        string
	ArgName     string
	TemplateKey string
	IsError     bool
}

type ErrorInfo struct {
	TypeName        string
	Fields          []ErrorField
	TextTemplate    string
	MessageTemplate string
	CodeValue       string
	HasErrorFields  bool
	ConstructorName string
	TextSprintf     string
	MessageSprintf  string
}

type FileInfo struct {
	PackageName string
	FileName    string
	Errors      []ErrorInfo
}

func main() {
	sourceFile := os.Getenv("GOFILE")
	if sourceFile == "" {
		fmt.Println("GOFILE environment variable is not set")
		os.Exit(1)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, sourceFile, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file: %v\n", err)
		os.Exit(1)
	}

	fileInfo := FileInfo{
		PackageName: node.Name.Name,
		FileName:    strings.TrimSuffix(sourceFile, ".go"),
	}

	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		if !hasErrorAnnotation(genDecl.Doc) {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			textTemplate, msgTemplate, codeValue := findErrorVariables(node, typeSpec.Name.Name)

			var fields []ErrorField
			hasErrorFields := false

			for _, field := range structType.Fields.List {
				fieldType := getTypeString(field.Type)
				isErrorField := fieldType == "error"

				if isErrorField {
					hasErrorFields = true
				}

				for _, name := range field.Names {
					fields = append(fields, ErrorField{
						Name:        name.Name,
						Type:        fieldType,
						ArgName:     stringcase.LowerCamel(name.Name),
						TemplateKey: name.Name,
						IsError:     isErrorField,
					})
				}
			}

			textSprintf := generateSprintf(textTemplate, fields)
			msgSprintf := generateSprintf(msgTemplate, fields)

			constructorName := "New" + stringcase.UpperCamel(typeSpec.Name.Name)

			fileInfo.Errors = append(fileInfo.Errors, ErrorInfo{
				TypeName:        typeSpec.Name.Name,
				ConstructorName: constructorName,
				Fields:          fields,
				TextTemplate:    textTemplate,
				MessageTemplate: msgTemplate,
				CodeValue:       codeValue,
				HasErrorFields:  hasErrorFields,
				TextSprintf:     textSprintf,
				MessageSprintf:  msgSprintf,
			})
		}
		return true
	})

	if len(fileInfo.Errors) == 0 {
		return
	}

	outputFile := fmt.Sprintf("%s_errors.gen.go", fileInfo.FileName)
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, fileInfo); err != nil {
		fmt.Printf("Failed to execute template: %v\n", err)
		os.Exit(1)
	}
}

func generateSprintf(templateStr string, fields []ErrorField) string {
	var sprintfFields []string

	if templateStr == "" {
		return ""
	}

	fieldMap := make(map[string]string)
	for _, field := range fields {
		fieldMap[field.Name] = field.Type
	}

	re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)

	matches := re.FindAllStringSubmatch(templateStr, -1)

	for _, match := range matches {
		field := match[1]

		if _, ok := fieldMap[field]; ok {
			sprintfFields = append(sprintfFields, field)
		}
	}

	args := []string{}
	for _, field := range sprintfFields {
		args = append(args, fmt.Sprintf("e.%s", field))
	}

	result := string(re.ReplaceAll([]byte(templateStr), []byte("%v")))

	if len(args) > 0 {
		sprintfArgs := strings.Join(args, ", ")
		return fmt.Sprintf(`fmt.Sprintf("%s", %s)`,
			strings.ReplaceAll(result, `"`, `\"`),
			sprintfArgs)
	}

	return `"` + result + `"`
}

func hasErrorAnnotation(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, comment := range doc.List {
		if strings.Contains(comment.Text, "@Error") {
			return true
		}
	}
	return false
}

func findErrorVariables(file *ast.File, typeName string) (text, message, code string) {
	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			return true
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) == 0 {
				continue
			}
			varName := valueSpec.Names[0].Name
			switch {
			case strings.HasSuffix(varName, "Text") && strings.Contains(strings.ToLower(varName), strings.ToLower(typeName)):
				if len(valueSpec.Values) > 0 {
					if lit, ok := valueSpec.Values[0].(*ast.BasicLit); ok {
						text = strings.Trim(lit.Value, `"`)
					}
				}
			case strings.HasSuffix(varName, "Message") && strings.Contains(strings.ToLower(varName), strings.ToLower(typeName)):
				if len(valueSpec.Values) > 0 {
					if lit, ok := valueSpec.Values[0].(*ast.BasicLit); ok {
						message = strings.Trim(lit.Value, `"`)
					}
				}
			case strings.HasSuffix(varName, "Code") && strings.Contains(strings.ToLower(varName), strings.ToLower(typeName)):
				if len(valueSpec.Values) > 0 {
					code = getValueString(valueSpec.Values[0])
				}
			}
		}
		return true
	})
	return text, message, code
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt)
		}
		return "[" + getTypeString(t.Len) + "]" + getTypeString(t.Elt)
	default:
		return "unknown"
	}
}

func getValueString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.BasicLit:
		return t.Value
	case *ast.SelectorExpr:
		return getValueString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}
