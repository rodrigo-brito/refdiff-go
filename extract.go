package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
	"text/scanner"
)

type NodeType string

const (
	FunctionType  NodeType = "Function"
	StructType    NodeType = "Struct"
	InterfaceType NodeType = "Interface"
)

type Node struct {
	Type           NodeType `json:"type"`
	Start          int      `json:"start"`
	End            int      `json:"end"`
	Name           string   `json:"name"`
	Tokens         []string `json:"tokens"`
	Namespace      string   `json:"namespace"`
	ParameterNames []string `json:"parameter_names,omitempty"`
	ParameterTypes []string `json:"parameter_types,omitempty"`
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

func (e Extractor) getTokens(start, end token.Pos) []string {
	var (
		scan   scanner.Scanner
		tokens []string
	)

	scan.Init(strings.NewReader(e.getSource(start, end)))
	scan.Filename = e.fileName
	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		tokens = append(tokens, scan.TokenText())
	}

	return tokens
}

func (e Extractor) getNamespace(sufix string) string {
	fileParts := strings.Split(e.fileName, "/")
	if len(fileParts) == 1 {
		return e.fileName
	}
	return strings.Join(fileParts[:len(fileParts)-1], "/")
}

func (e Extractor) Extract() []Node {
	nodes := make([]Node, 0)
	namesapace := e.getNamespace("")

	ast.Inspect(e.astFile, func(node ast.Node) bool {
		switch definition := node.(type) {
		case *ast.TypeSpec: // Structs / Interfaces
			if structType, ok := definition.Type.(*ast.StructType); ok {
				nodes = append(nodes, Node{
					Start:     e.fileSet.Position(structType.Pos()).Line,
					End:       e.fileSet.Position(structType.End()).Line,
					Name:      definition.Name.Name,
					Type:      StructType,
					Namespace: namesapace,
					Tokens:    e.getTokens(structType.Pos(), structType.End()),
				})
			} else if iface, ok := definition.Type.(*ast.InterfaceType); ok {
				nodes = append(nodes, Node{
					Start:     e.fileSet.Position(iface.Pos()).Line,
					End:       e.fileSet.Position(iface.End()).Line,
					Name:      definition.Name.Name,
					Type:      InterfaceType,
					Namespace: namesapace,
					Tokens:    e.getTokens(iface.Pos(), iface.End()),
				})
			}

		case *ast.FuncDecl: // Functions
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

			name := definition.Name.Name
			if structName != "" {
				name = fmt.Sprintf("%s.%s", structName, name)
			}

			paramTypes := make([]string, 0)
			paramNames := make([]string, 0)
			if definition.Type.Params != nil {
				for _, field := range definition.Type.Params.List {
					if len(field.Names) > 0 {
						paramNames = append(paramNames, field.Names[0].Name)
						paramTypes = append(paramTypes, e.getSource(field.Type.Pos(), field.Type.End()))
					}
				}
			}

			nodes = append(nodes, Node{
				Start:          e.fileSet.Position(definition.Body.Pos()).Line,
				End:            e.fileSet.Position(definition.Body.End()).Line,
				Name:           name,
				ParameterNames: paramNames,
				ParameterTypes: paramTypes,
				Type:           FunctionType,
				Namespace:      namesapace,
				Tokens:         e.getTokens(definition.Body.Pos(), definition.Body.End()),
			})
		}
		return true
	})

	return nodes
}
