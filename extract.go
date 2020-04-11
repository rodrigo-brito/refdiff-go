package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"strings"
)

type NodeType string

const (
	FileType      NodeType = "File"
	StructType    NodeType = "Struct"
	InterfaceType NodeType = "Interface"
	FunctionType  NodeType = "Function"
)

type Node struct {
	Type           NodeType `json:"type"`
	Start          int      `json:"start"`
	End            int      `json:"end"`
	Name           string   `json:"name"`
	Namespace      string   `json:"namespace"`
	Parent         *string  `json:"parent"`
	Tokens         []string `json:"tokens,omitempty"`
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

func (e Extractor) getFileTokens(start, end token.Pos) []string {
	var (
		scan   scanner.Scanner
		tokens []string
	)

	fset := token.NewFileSet()
	source := e.getSource(start, end)
	scan.Init(fset.AddFile(e.fileName, fset.Base(), len(source)), []byte(source), nil, scanner.ScanComments)
	for {
		position, tok, literal := scan.Scan()
		if tok == token.EOF {
			break
		}

		if literal == "" {
			literal = tok.String()
		}

		tokens = append(tokens, fmt.Sprintf("%d-%d", position, int(position)+len(literal)+1))
	}

	return tokens
}

func (e Extractor) getShortFilename() string {
	parts := strings.Split(e.fileName, "/")
	return parts[len(parts)-1]
}

func (e Extractor) getNamespace(sufix string) string {
	fileParts := strings.Split(e.fileName, "/")
	if len(fileParts) == 1 {
		return ""
	}
	namespace := strings.Join(fileParts[:len(fileParts)-1], "/")
	if len(sufix) > 0 {
		namespace = fmt.Sprintf("%s/%s.", namespace, sufix)
	} else {
		namespace = fmt.Sprintf("%s/", namespace)
	}
	return namespace
}

func (e Extractor) Extract() []Node {
	nodes := make([]Node, 0)

	ast.Inspect(e.astFile, func(node ast.Node) bool {
		switch definition := node.(type) {
		case *ast.File:
			nodes = append(nodes, Node{
				Start:     e.fileSet.Position(definition.Pos()).Line,
				End:       e.fileSet.Position(definition.End()).Line,
				Name:      e.getShortFilename(),
				Type:      FileType,
				Namespace: e.getNamespace(""),
				Tokens:    e.getFileTokens(definition.Pos(), definition.End()),
			})
		case *ast.TypeSpec: // Structs / Interfaces
			parent := e.getNamespace("") + e.getShortFilename()
			if structType, ok := definition.Type.(*ast.StructType); ok {
				nodes = append(nodes, Node{
					Type:      StructType,
					Parent:    &parent,
					Namespace: e.getNamespace(""),
					Name:      definition.Name.Name,
					Start:     e.fileSet.Position(structType.Pos()).Line,
					End:       e.fileSet.Position(structType.End()).Line,
				})
			} else if iface, ok := definition.Type.(*ast.InterfaceType); ok {
				nodes = append(nodes, Node{
					Type:   InterfaceType,
					Parent: &parent,
					Name:   definition.Name.Name,
					Start:  e.fileSet.Position(iface.Pos()).Line,
					End:    e.fileSet.Position(iface.End()).Line,
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

			parent := e.getNamespace("") + e.getShortFilename()
			if len(structName) > 0 {
				parent = e.getNamespace("") + structName
			}
			nodes = append(nodes, Node{
				Start:          e.fileSet.Position(definition.Body.Pos()).Line,
				End:            e.fileSet.Position(definition.Body.End()).Line,
				Name:           definition.Name.Name,
				ParameterNames: paramNames,
				ParameterTypes: paramTypes,
				Type:           FunctionType,
				Parent:         &parent,
				Namespace:      e.getNamespace(structName),
			})
		}
		return true
	})

	return nodes
}
