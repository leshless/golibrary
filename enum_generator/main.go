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

//go:embed enum.go.tpl
var enumFile string
var enumFileTemplate = template.Must(template.New("enum").Parse(enumFile))

var (
	enumRegexp = regexp.MustCompile("@Enum")
)

type ConstantCriteria struct {
	Name        string
	StringValue string
}

type EnumCriteria struct {
	TypeName            string
	Constants           []ConstantCriteria
	DefaultConstantName string
}

type FileCriteria struct {
	PackageName string
	FileName    string
	Enums       []EnumCriteria
}

func main() {
	sourceFileName := os.Getenv("GOFILE")

	fset := token.NewFileSet()
	sourceFile, err := parser.ParseFile(fset, sourceFileName, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file source file: %v\n", err)
		os.Exit(1)
	}

	enums := make([]EnumCriteria, 0)
	ast.Inspect(sourceFile, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		if genDecl.Doc == nil {
			return false
		}

		shouldOmitt := true

		for _, comment := range genDecl.Doc.List {
			if enumRegexp.MatchString(comment.Text) {
				shouldOmitt = false
				break
			}
		}

		if shouldOmitt {
			return false
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			constants := getConstants(sourceFile, typeSpec.Name.Name)
			if len(constants) == 0 {
				continue
			}

			enums = append(enums, EnumCriteria{
				TypeName:            typeSpec.Name.Name,
				Constants:           constants,
				DefaultConstantName: constants[0].Name,
			})
		}
		return true
	})

	if len(enums) == 0 {
		return
	}

	fileCriteria := FileCriteria{
		PackageName: sourceFile.Name.Name,
		FileName:    strings.TrimSuffix(sourceFileName, ".go"),
		Enums:       enums,
	}

	outputFile := fmt.Sprintf("%s_enums.gen.go", fileCriteria.FileName)
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := enumFileTemplate.Execute(f, fileCriteria); err != nil {
		fmt.Printf("Failed to execute template: %v\n", err)
		os.Exit(1)
	}
}

func getConstants(file *ast.File, typeName string) []ConstantCriteria {
	constants := make([]ConstantCriteria, 0)
	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			return true
		}

		for i, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			if i == 0 && valueSpec.Type == nil {
				break
			}

			ident, ok := valueSpec.Type.(*ast.Ident)
			if ok {
				if ident.Name != typeName {
					break
				}
			}

			for _, name := range valueSpec.Names {
				stringValue := stringcase.UpperSnake(strings.TrimPrefix(name.Name, typeName))

				constants = append(constants, ConstantCriteria{
					Name:        name.Name,
					StringValue: stringValue,
				})
			}
		}
		return true
	})

	return constants
}
