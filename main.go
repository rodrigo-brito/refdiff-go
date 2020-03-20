package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type NodeType string

const (
	MethodType NodeType = "METHOD"
	StructType NodeType = "STRUCT"
)

type Node struct {
	Start          int
	End            int
	Body           string
	Identification string
	Type           NodeType
}

func (n Node) String() string {
	return fmt.Sprintf("[%d-%d] %s - %s", n.Start, n.End, n.Type, n.Identification)
}

func main() {
	var fileName = flag.String("file", "", "file path, ex: main.go")
	flag.NewFlagSet("file", flag.ExitOnError)
	flag.Parse()

	if *fileName == "" {
		fmt.Println("flag -file required.")
		os.Exit(2)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, *fileName, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	nodes := make([]Node, 0)

	ast.Inspect(file, func(node ast.Node) bool {
		switch definition := node.(type) {
		case *ast.TypeSpec: // Structs
			if structObj, ok := definition.Type.(*ast.StructType); ok {
				var paramTypes []string
				for _, field := range structObj.Fields.List {
					paramTypes = append(paramTypes, fmt.Sprint(field.Type))
				}

				nodes = append(nodes, Node{
					Start:          fset.Position(structObj.Pos()).Line,
					End:            fset.Position(structObj.End()).Line,
					Body:           "",
					Identification: definition.Name.Name,
					Type:           StructType,
				})
			}
		case *ast.FuncDecl: // Methods
			var structName string
			if definition.Recv != nil && len(definition.Recv.List) > 0 {
				typeObj := definition.Recv.List[0].Type
				if ident, ok := typeObj.(*ast.Ident); ok {
					structName = ident.Name
				} else {
					if ident, ok := typeObj.(*ast.StarExpr).X.(*ast.Ident); ok {
						structName = ident.Name
					} else if ident, ok := typeObj.(*ast.StarExpr).X.(*ast.SelectorExpr); ok {
						structName = ident.Sel.Name
					}
				}
			}

			var paramsType []string
			if definition.Recv != nil {
				for _, field := range definition.Recv.List {
					paramsType = append(paramsType, fmt.Sprint(field.Type))
				}
			}

			name := fmt.Sprintf("%s(%s)", definition.Name.Name, strings.Join(paramsType, ","))
			if len(structName) > 0 {
				name = fmt.Sprintf("(%s) %s", structName, name)
			}

			nodes = append(nodes, Node{
				Start:          fset.Position(definition.Body.Pos()).Line,
				End:            fset.Position(definition.Body.End()).Line,
				Body:           "",
				Identification: name,
				Type:           MethodType,
			})
		}
		return true
	})

	for _, node := range nodes {
		fmt.Println(node)
	}
}
