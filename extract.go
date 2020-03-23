package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

type NodeType string

const (
	MethodType NodeType = "METHOD"
	StructType NodeType = "STRUCT"
)

type Node struct {
	Start          int      `json:"start"`
	End            int      `json:"end"`
	Body           string   `json:"body"`
	Name           string   `json:"name"`
	Type           NodeType `json:"type"`
	Namespace      string   `json:"namespace"`
	ParameterNames []string `json:"parameter_names"`
	ParameterTypes []string `json:"parameter_types"`
}

func (n Node) String() string {
	return fmt.Sprintf("[%d-%d] %s - %s", n.Start, n.End, n.Type, n.Name)
}

type Extractor struct {
	fileName    string
	fileContent []byte
	astFile     *ast.File
	fileSet     *token.FileSet
}

func NewExtractor(filename string) (*Extractor, error) {
	extractor := &Extractor{
		fileName: filename,
		fileSet:  token.NewFileSet(),
	}

	var err error
	extractor.fileContent, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	extractor.astFile, err = parser.ParseFile(extractor.fileSet, filename, extractor.fileContent, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return extractor, nil
}

func (e Extractor) getSource(start, end token.Pos) string {
	return string(e.fileContent[start-e.astFile.Pos() : end-e.astFile.Pos()])
}

func (e Extractor) Extract() []Node {
	nodes := make([]Node, 0)

	ast.Inspect(e.astFile, func(node ast.Node) bool {
		switch definition := node.(type) {
		case *ast.TypeSpec: // Structs
			if structType, ok := definition.Type.(*ast.StructType); ok {
				nodes = append(nodes, Node{
					Start: e.fileSet.Position(structType.Pos()).Line,
					End:   e.fileSet.Position(structType.End()).Line,
					Body:  e.getSource(structType.Pos(), structType.End()),
					Name:  definition.Name.Name,
					Type:  StructType,
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

			var paramTypes []string
			var paramNames []string
			if definition.Type.Params != nil {
				for _, field := range definition.Type.Params.List {
					if len(field.Names) > 0 {
						paramNames = append(paramNames, field.Names[0].Name)
						paramTypes = append(paramTypes, e.getSource(field.Type.Pos(), field.Type.End()))
					}
				}
			}

			name := fmt.Sprintf("%s(%s)", definition.Name.Name, strings.Join(paramTypes, ","))
			if len(structName) > 0 {
				name = fmt.Sprintf("(%s) %s", structName, name)
			}

			nodes = append(nodes, Node{
				Start:          e.fileSet.Position(definition.Body.Pos()).Line,
				End:            e.fileSet.Position(definition.Body.End()).Line,
				Name:           name,
				Body:           e.getSource(definition.Body.Pos(), definition.Body.End()),
				Type:           MethodType,
				ParameterNames: paramNames,
				ParameterTypes: paramTypes,
			})
		}
		return true
	})

	return nodes
}
