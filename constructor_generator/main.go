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

//go:embed constructor.go.tpl
var constructorFile string
var constructorFileTemplate = template.Must(template.New("constructor").Parse(constructorFile))

var (
	valueInstanceRegexp   = regexp.MustCompile("@[a-zA-Z]*ValueInstance")
	pointerInstanceRegexp = regexp.MustCompile("@[a-zA-Z]*PointerInstance")
	publicInstanceRegexp  = regexp.MustCompile("@Public[a-zA-Z]*Instance")
	privateInstanceRegexp = regexp.MustCompile("@Private[a-zA-Z]*Instance")
)

type constructorCriteria struct {
	FileName        string
	PackageName     string
	StructName      string
	ConstructorName string
	IsValueInstance bool
	Fields          []fieldInfo
	Imports         []importInfo
}

type fieldInfo struct {
	Name    string
	Type    string
	ArgName string
}

type importInfo struct {
	Name string
	Path string
}

func main() {
	sourceFileName := os.Getenv("GOFILE")

	sourceFile, err := parser.ParseFile(token.NewFileSet(), sourceFileName, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse source file: %v\n", err)
		os.Exit(1)
	}

	criterias := make([]constructorCriteria, 0)
	allImports := parseImports(sourceFile)

	ast.Inspect(sourceFile, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.GenDecl:
			if node.Tok != token.TYPE {
				return true
			}

			for _, spec := range node.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Check for constructor annotations in comments
				var (
					shouldOmitt      bool
					isValueInstance  bool
					isPublicInstance bool
				)
				if node.Doc != nil {
					if len(node.Doc.List) == 0 {
						continue
					}

					for _, comment := range node.Doc.List {
						text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

						if valueInstanceRegexp.Match([]byte(text)) {
							isValueInstance = true
						} else if !pointerInstanceRegexp.Match([]byte(text)) {
							shouldOmitt = true
							break
						}

						if publicInstanceRegexp.Match([]byte(text)) {
							isPublicInstance = true
						} else if !privateInstanceRegexp.Match([]byte(text)) {
							shouldOmitt = true
							break
						}
					}
				}

				if shouldOmitt {
					continue
				}

				fields := parseStructFields(structType, sourceFile)
				imports := filterUsedImportsForFields(fields, allImports)

				var constructorName string
				if isPublicInstance {
					constructorName = "New" + stringcase.UpperCamel(typeSpec.Name.Name)
				} else {
					constructorName = "new" + stringcase.UpperCamel(typeSpec.Name.Name)
				}

				criterias = append(criterias, constructorCriteria{
					FileName:        sourceFileName,
					PackageName:     sourceFile.Name.Name,
					StructName:      typeSpec.Name.Name,
					ConstructorName: constructorName,
					IsValueInstance: isValueInstance,
					Fields:          fields,
					Imports:         imports,
				})
			}
		}
		return true
	})

	if len(criterias) == 0 {
		return
	}

	for _, criteria := range criterias {
		err = createConstructorFileFromCriteria(criteria)
		if err != nil {
			fmt.Printf("Failed to create constructor file for struct %s: %v\n", criteria.StructName, err)
			os.Exit(1)
		}
	}
}

// filterUsedImportsForFields returns only imports that are used in the specific struct fields
func filterUsedImportsForFields(fields []fieldInfo, allImports []importInfo) []importInfo {
	usedImports := make(map[string]importInfo)

	// Create a map for quick lookup of imports by name
	importsByName := make(map[string]importInfo)
	for _, imp := range allImports {
		importsByName[imp.Name] = imp
	}

	// Check each field type for used imports
	for _, field := range fields {
		used := findUsedImportsInType(field.Type, importsByName)
		for name, imp := range used {
			usedImports[name] = imp
		}
	}

	// Convert map back to slice
	var result []importInfo
	for _, imp := range usedImports {
		result = append(result, imp)
	}

	fmt.Printf("Filtered imports for struct: %d used out of %d total\n", len(result), len(allImports))
	return result
}

// findUsedImportsInType recursively checks a type string for used imports
func findUsedImportsInType(typeStr string, importsByName map[string]importInfo) map[string]importInfo {
	used := make(map[string]importInfo)

	// Simple check: if the type contains "importName." then that import is used
	for importName, imp := range importsByName {
		if strings.Contains(typeStr, importName+".") {
			used[importName] = imp
		}
	}

	return used
}

func parseImports(file *ast.File) []importInfo {
	var imports []importInfo

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		var importName string
		if imp.Name != nil {
			importName = imp.Name.Name
		} else {
			// Extract the last part of the path as the default name
			parts := strings.Split(importPath, "/")
			importName = parts[len(parts)-1]
		}

		imports = append(imports, importInfo{
			Name: importName,
			Path: importPath,
		})
	}

	return imports
}

func parseStructFields(structType *ast.StructType, file *ast.File) []fieldInfo {
	var fields []fieldInfo

	for _, field := range structType.Fields.List {
		// Skip fields without names (embedded fields)
		if len(field.Names) == 0 {
			continue
		}

		for _, name := range field.Names {
			typeString := getTypeString(field.Type, file)
			argName := stringcase.LowerCamel(name.Name)

			fields = append(fields, fieldInfo{
				Name:    name.Name,
				Type:    typeString,
				ArgName: argName,
			})
		}
	}

	return fields
}

func getTypeString(expr ast.Expr, file *ast.File) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		pkg := getTypeString(t.X, file)
		return pkg + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X, file)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt, file)
		}
		return "[" + getTypeString(t.Len, file) + "]" + getTypeString(t.Elt, file)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key, file) + "]" + getTypeString(t.Value, file)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.ChanType:
		dir := ""
		switch t.Dir {
		case ast.SEND:
			dir = "chan<- "
		case ast.RECV:
			dir = "<-chan "
		default:
			dir = "chan "
		}
		return dir + getTypeString(t.Value, file)
	case *ast.FuncType:
		return "func" + getFuncTypeString(t)
	default:
		return "unknown"
	}
}

func getFuncTypeString(funcType *ast.FuncType) string {
	var params, results []string

	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			typeStr := getTypeString(param.Type, nil)
			if len(param.Names) > 0 {
				for range param.Names {
					params = append(params, typeStr)
				}
			} else {
				params = append(params, typeStr)
			}
		}
	}

	if funcType.Results != nil {
		for _, result := range funcType.Results.List {
			typeStr := getTypeString(result.Type, nil)
			if len(result.Names) > 0 {
				for range result.Names {
					results = append(results, typeStr)
				}
			} else {
				results = append(results, typeStr)
			}
		}
	}

	result := "(" + strings.Join(params, ", ") + ")"
	if len(results) > 0 {
		if len(results) == 1 {
			result += " " + results[0]
		} else {
			result += " (" + strings.Join(results, ", ") + ")"
		}
	}

	return result
}

func createConstructorFileFromCriteria(criteria constructorCriteria) error {
	fileName := fmt.Sprintf("new_%s.gen.go", stringcase.LowerSnake(criteria.StructName))

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() { _ = file.Close() }()

	err = constructorFileTemplate.Execute(file, criteria)
	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}
