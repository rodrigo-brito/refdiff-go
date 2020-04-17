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
	Line           int      `json:"line"`
	HasBody        bool     `json:"has_body"`
	Name           string   `json:"name"`
	Namespace      string   `json:"namespace"`
	Parent         *string  `json:"parent"`
	Receiver       *string  `json:"receiver"`
	Tokens         []string `json:"tokens,omitempty"`
	ParameterNames []string `json:"parameter_names,omitempty"`
	ParameterTypes []string `json:"parameter_types,omitempty"`
	FunctionCalls  []string `json:"function_calls,omitempty"`
}

func (n Node) String() string {
	return fmt.Sprintf("[%d-%d] %s - %s", n.Start, n.End, n.Type, n.Name)
}

type Extractor struct {
	rootFolder    string
	fileName      string
	fileContent   []byte
	astFile       *ast.File
	fileSet       *token.FileSet
	functionCalls map[string][]string
}

func NewExtractor(rootFolder, filename string) (*Extractor, error) {
	var err error
	extractor := &Extractor{
		rootFolder: rootFolder,
		fileName:   filename,
		fileSet:    token.NewFileSet(),
	}

	extractor.functionCalls, err = GetCallGraph(rootFolder, filename)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/%s", rootFolder, filename)
	extractor.fileContent, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	extractor.astFile, err = parser.ParseFile(extractor.fileSet, path, extractor.fileContent, parser.ParseComments)
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
	scan.Init(fset.AddFile(
		fmt.Sprintf("%s/%s", e.rootFolder, e.fileName),
		fset.Base(),
		len(source),
	), []byte(source), nil, scanner.ScanComments)
	for {
		position, tok, literal := scan.Scan()
		if tok == token.EOF {
			break
		}

		if literal == "" {
			literal = tok.String()
		}

		tokens = append(tokens, fmt.Sprintf("%d-%d", position-1, int(position-1)+len(literal)))
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

func (e Extractor) extractParameters(fieldList *ast.FieldList) (names []string, types []string) {
	paramTypes := make([]string, 0)
	paramNames := make([]string, 0)
	if fieldList != nil {
		for index, field := range fieldList.List {
			paramTypes = append(paramTypes, e.getSource(field.Type.Pos(), field.Type.End()))
			if len(field.Names) > 0 {
				paramNames = append(paramNames, field.Names[0].Name)
			} else {
				paramNames = append(paramNames, fmt.Sprintf("p%d", index+1))
			}
		}
	}
	return paramNames, paramTypes
}

func (e Extractor) Extract() []Node {
	nodes := make([]Node, 0)

	ast.Inspect(e.astFile, func(node ast.Node) bool {
		switch definition := node.(type) {
		case *ast.File:
			nodes = append(nodes, Node{
				Start:     int(definition.Pos()) - 1,
				End:       int(definition.End()) - 1,
				Line:      e.fileSet.Position(definition.Pos()).Line,
				HasBody:   true,
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
					Start:     int(structType.Pos()) - 1,
					End:       int(structType.End()) - 1,
					Line:      e.fileSet.Position(structType.Pos()).Line,
					HasBody:   true,
				})
			} else if iface, ok := definition.Type.(*ast.InterfaceType); ok {
				interfaceName := definition.Name.Name
				nodes = append(nodes, Node{
					Type:      InterfaceType,
					Parent:    &parent,
					Namespace: e.getNamespace(""),
					Name:      interfaceName,
					Start:     int(iface.Pos()) - 1,
					End:       int(iface.End()) - 1,
					Line:      e.fileSet.Position(iface.Pos()).Line,
					HasBody:   false,
				})

				for _, definition := range iface.Methods.List {
					if function, ok := definition.Type.(*ast.FuncType); ok {
						paramNames, paramTypes := e.extractParameters(function.Params)
						parent := e.getNamespace("") + interfaceName
						nodes = append(nodes, Node{
							Type:           FunctionType,
							Name:           definition.Names[0].Name,
							Start:          int(definition.Pos()) - 1,
							End:            int(definition.End()) - 1,
							Line:           e.fileSet.Position(definition.Pos()).Line,
							Parent:         &parent,
							Namespace:      e.getNamespace(interfaceName),
							ParameterNames: paramNames,
							ParameterTypes: paramTypes,
							HasBody:        false,
						})
					}
				}
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

			paramNames, paramTypes := e.extractParameters(definition.Type.Params)
			parent := e.getNamespace("") + e.getShortFilename()
			if len(structName) > 0 {
				parent = e.getNamespace("") + structName
			}
			nodes = append(nodes, Node{
				Type:           FunctionType,
				Name:           definition.Name.Name,
				Parent:         &parent,
				Namespace:      e.getNamespace(structName),
				ParameterNames: paramNames,
				ParameterTypes: paramTypes,
				Receiver:       &structName,
				Start:          int(definition.Pos()) - 1,
				End:            int(definition.End()) - 1,
				Line:           e.fileSet.Position(definition.Pos()).Line,
				HasBody:        true,
				FunctionCalls:  e.functionCalls[e.getNamespace(structName)+definition.Name.Name],
			})
		}
		return true
	})

	return nodes
}
