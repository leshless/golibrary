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

	"github.com/leshless/golibrary/set"
	"github.com/leshless/golibrary/sets"
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

type importCriteria struct {
	Name string
	Path string
}

type fieldCriteria struct {
	Name    string
	Type    string
	ArgName string
}

type constructorCriteria struct {
	StructName      string
	ConstructorName string
	IsValueInstance bool
	Fields          []fieldCriteria
}

type fileCriteria struct {
	FileName     string
	PackageName  string
	Imports      []importCriteria
	Constructors []constructorCriteria
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
	usedImports := set.New[importCriteria]()

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

				if node.Doc == nil {
					continue
				}

				if len(node.Doc.List) == 0 {
					continue
				}

				// Check for constructor annotations in comments
				var (
					shouldOmitt      bool
					isValueInstance  bool
					isPublicInstance bool
				)

				for _, comment := range node.Doc.List {
					text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

					if valueInstanceRegexp.MatchString(text) {
						isValueInstance = true
					} else if !pointerInstanceRegexp.MatchString(text) {
						shouldOmitt = true
						break
					}

					if publicInstanceRegexp.MatchString(text) {
						isPublicInstance = true
					} else if !privateInstanceRegexp.Match([]byte(text)) {
						shouldOmitt = true
						break
					}
				}

				if shouldOmitt {
					continue
				}

				fields := parseStructFields(structType)
				imports := filterUsedImportsForFields(fields, allImports)

				usedImports = sets.Union(usedImports, set.FromSlice(imports))

				var constructorName string
				if isPublicInstance {
					constructorName = "New" + stringcase.UpperCamel(typeSpec.Name.Name)
				} else {
					constructorName = "new" + stringcase.UpperCamel(typeSpec.Name.Name)
				}

				criterias = append(criterias, constructorCriteria{
					StructName:      typeSpec.Name.Name,
					ConstructorName: constructorName,
					IsValueInstance: isValueInstance,
					Fields:          fields,
				})
			}
		}
		return true
	})

	if len(criterias) == 0 {
		return
	}

	criteria := fileCriteria{
		FileName:     strings.Split(sourceFileName, ".")[0],
		PackageName:  sourceFile.Name.Name,
		Imports:      usedImports.Slice(),
		Constructors: criterias,
	}

	err = createConstructorFileFromCriteria(criteria)
	if err != nil {
		fmt.Printf("Failed to create constructor file: %s\n", err.Error())
		os.Exit(1)
	}
}

// filterUsedImportsForFields returns only imports that are used in the specific struct fields
func filterUsedImportsForFields(fields []fieldCriteria, allImports []importCriteria) []importCriteria {
	usedImports := make(map[string]importCriteria)

	// Create a map for quick lookup of imports by name
	importsByName := make(map[string]importCriteria)
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
	var result []importCriteria
	for _, imp := range usedImports {
		result = append(result, imp)
	}

	return result
}

// findUsedImportsInType recursively checks a type string for used imports
func findUsedImportsInType(typeStr string, importsByName map[string]importCriteria) map[string]importCriteria {
	used := make(map[string]importCriteria)

	// Simple check: if the type contains "importName." then that import is used
	for importName, imp := range importsByName {
		if strings.Contains(typeStr, importName+".") {
			used[importName] = imp
		}
	}

	return used
}

func parseImports(file *ast.File) []importCriteria {
	var imports []importCriteria

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

		imports = append(imports, importCriteria{
			Name: importName,
			Path: importPath,
		})
	}

	return imports
}

func parseStructFields(structType *ast.StructType) []fieldCriteria {
	var fields []fieldCriteria

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			typeString := getTypeString(field.Type)
			typeStringPath := strings.Split(typeString, ".")

			var name string
			if len(typeStringPath) == 2 {
				name = typeStringPath[1]
			} else {
				name = typeStringPath[0]
			}

			fields = append(fields, fieldCriteria{
				Name:    name,
				Type:    typeString,
				ArgName: stringcase.LowerCamel(name),
			})
		}

		for _, name := range field.Names {
			typeString := getTypeString(field.Type)

			fields = append(fields, fieldCriteria{
				Name:    name.Name,
				Type:    typeString,
				ArgName: stringcase.LowerCamel(name.Name),
			})
		}
	}

	return fields
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		pkg := getTypeString(t.X)
		return pkg + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt)
		}
		return "[" + getTypeString(t.Len) + "]" + getTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
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
		return dir + getTypeString(t.Value)
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
			typeStr := getTypeString(param.Type)
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
			typeStr := getTypeString(result.Type)
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

func createConstructorFileFromCriteria(criteria fileCriteria) error {
	fileName := fmt.Sprintf("%s_constructors.gen.go", stringcase.LowerSnake(criteria.FileName))

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
