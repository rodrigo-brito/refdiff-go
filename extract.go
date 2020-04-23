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
	PrimitiveType NodeType = "Type"
	FromRootFile           = "_ROOT_"
)

type Node struct {
	Type      NodeType `json:"type"`
	Start     int      `json:"start"`
	End       int      `json:"end"`
	Line      int      `json:"line"`
	HasBody   bool     `json:"has_body"`
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Parent    *string  `json:"parent"`
	Tokens    []string `json:"tokens,omitempty"`

	// only for functions
	FunctionCalls  []string `json:"function_calls,omitempty"`
	ParameterNames []string `json:"parameter_names,omitempty"`
	ParameterTypes []string `json:"parameter_types,omitempty"`
	Receiver       *string  `json:"receiver"`
	ReceiverAlias  *string  `json:"-"`
}

func (n Node) String() string {
	return fmt.Sprintf("[%d-%d] %s - %s", n.Start, n.End, n.Type, n.Name)
}

type Extractor struct {
	rootFolder  string
	fileName    string
	fileContent []byte
	astFile     *ast.File
	fileSet     *token.FileSet
}

type FunctionCall struct {
	Position int
	Prefix   string
	Suffix   string
	Done     bool
}

func NewExtractor(rootFolder, filename string) (*Extractor, error) {
	var err error
	extractor := &Extractor{
		rootFolder: rootFolder,
		fileName:   filename,
		fileSet:    token.NewFileSet(),
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
	startPos := e.fileSet.Position(start).Offset
	endPos := e.fileSet.Position(end).Offset
	return string(e.fileContent[startPos:endPos])
}

func (e Extractor) getFileTokens() []string {
	var (
		scan   scanner.Scanner
		tokens []string
	)

	fset := token.NewFileSet()
	scan.Init(fset.AddFile(
		fmt.Sprintf("%s/%s", e.rootFolder, e.fileName),
		fset.Base(),
		len(e.fileContent),
	), e.fileContent, nil, scanner.ScanComments)
	for {
		position, tok, literal := scan.Scan()
		if tok == token.EOF {
			break
		}

		if literal == "" {
			literal = tok.String()
		}

		pos := fset.Position(position)
		tokens = append(tokens, fmt.Sprintf("%d:%d:%d", pos.Line-1, pos.Column-1, len(literal)))
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
		if len(sufix) > 0 {
			return sufix + "."
		}
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
			typeName := e.getSource(field.Type.Pos(), field.Type.End())
			if len(field.Names) > 0 {
				for _, nameObject := range field.Names {
					paramTypes = append(paramTypes, typeName)
					paramNames = append(paramNames, nameObject.Name)
				}
			} else {
				paramTypes = append(paramTypes, typeName)
				paramNames = append(paramNames, fmt.Sprintf("p%d", index+1))
			}
		}
	}
	return paramNames, paramTypes
}

func (e Extractor) Extract() []*Node {
	nodes := make([]*Node, 0)
	functionCalls := make([]*FunctionCall, 0)

	ast.Inspect(e.astFile, func(node ast.Node) bool {
		switch definition := node.(type) {
		case *ast.File:
			nodes = append(nodes, &Node{
				Type:      FileType,
				Name:      e.getShortFilename(),
				Namespace: e.getNamespace(""),
				Start:     int(definition.Pos()) - 1,
				End:       int(definition.End()) - 1,
				Line:      e.fileSet.Position(definition.Pos()).Line,
				HasBody:   true,
				Tokens:    e.getFileTokens(),
			})
		case *ast.TypeSpec: // Structs / Interfaces
			parent := e.getNamespace("") + e.getShortFilename()
			if structType, ok := definition.Type.(*ast.StructType); ok {
				nodes = append(nodes, &Node{
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
				nodes = append(nodes, &Node{
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
						nodes = append(nodes, &Node{
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
			var receiverName string
			var receiverAlias string
			if definition.Recv != nil && len(definition.Recv.List) > 0 {
				receiverName = Ident(definition.Recv.List[0].Type)
				if len(definition.Recv.List[0].Names) > 0 {
					receiverAlias = Ident(definition.Recv.List[0].Names[0])
				}
			}

			paramNames, paramTypes := e.extractParameters(definition.Type.Params)
			parent := e.getNamespace("") + e.getShortFilename()
			if len(receiverName) > 0 {
				parent = e.getNamespace("") + receiverName
			}
			nodes = append(nodes, &Node{
				Type:           FunctionType,
				Name:           definition.Name.Name,
				Parent:         &parent,
				Namespace:      e.getNamespace(receiverName),
				ParameterNames: paramNames,
				ParameterTypes: paramTypes,
				Receiver:       &receiverName,
				ReceiverAlias:  &receiverAlias,
				Start:          int(definition.Pos()) - 1,
				End:            int(definition.End()) - 1,
				Line:           e.fileSet.Position(definition.Pos()).Line,
				HasBody:        true,
			})
		case *ast.GenDecl:
			if len(definition.Specs) > 0 {
				if spec, ok := definition.Specs[0].(*ast.TypeSpec); ok {
					switch spec.Type.(type) {
					case *ast.Ident,
						*ast.MapType,
						*ast.ArrayType,
						*ast.SelectorExpr:
						parent := e.getNamespace("") + e.getShortFilename()
						nodes = append(nodes, &Node{
							Type:      PrimitiveType,
							Name:      spec.Name.Name,
							Parent:    &parent,
							Namespace: e.getNamespace(""),
							Start:     int(definition.Pos()) - 1,
							End:       int(definition.End()) - 1,
							Line:      e.fileSet.Position(definition.Pos()).Line,
							HasBody:   false,
						})
					}
				}
			}

		case *ast.CallExpr: // Function calls
			ident, ok := definition.Fun.(*ast.Ident)
			if ok && ident.Obj != nil {
				functionCalls = append(functionCalls, &FunctionCall{
					Position: int(definition.Pos()) - 1,
					Prefix:   FromRootFile,
					Suffix:   ident.Obj.Name,
				})

			} else if sel, ok := definition.Fun.(*ast.SelectorExpr); ok {
				functionCalls = append(functionCalls, &FunctionCall{
					Position: int(definition.Pos()) - 1,
					Prefix:   Ident(sel.X),
					Suffix:   Ident(sel.Sel),
				})
			}
		}
		return true
	})

	// get all function calls detected and insert in the origin nodes
	e.updateFunctionCalls(nodes, functionCalls)

	return nodes
}

func (e Extractor) updateFunctionCalls(nodes []*Node, functionCalls []*FunctionCall) {
	for _, node := range nodes {
		if node.Type != FunctionType {
			continue
		}

		for _, call := range functionCalls {
			if call.Done || call.Position < node.Start || call.Position > node.End {
				continue
			}

			// Local method
			if call.Prefix == FromRootFile {
				node.FunctionCalls = append(node.FunctionCalls, e.getNamespace("")+call.Suffix)
			} else if call.Prefix != "" && node.ReceiverAlias != nil && *node.ReceiverAlias == call.Prefix {
				node.FunctionCalls = append(node.FunctionCalls, fmt.Sprintf("%s.%s", *node.Parent, call.Suffix))
			}
			call.Done = true
		}
	}
}

func ObjectName(obj *ast.Object) (pkg string, typ string) {
	switch decl := obj.Decl.(type) {
	case *ast.Field:
		return Selector(decl.Type)
	case *ast.AssignStmt:
		idx := -1
		for ii, v := range decl.Lhs {
			if Ident(v) == obj.Name {
				idx = ii
				break
			}
		}
		if idx >= 0 {
			return Selector(decl.Rhs[idx])
		}
	}
	return "", ""
}

func Ident(node ast.Expr) string {
	switch n := node.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.SelectorExpr:
		x := Ident(n.X)
		s := Ident(n.Sel)
		if x != "" && s != "" {
			return x + "." + s
		}
		return s
	case *ast.StarExpr: // *StructName
		return Ident(n.X)
	}
	return ""
}

func Selector(expr ast.Expr) (x string, sel string) {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return Selector(e.X)
	case *ast.UnaryExpr:
		return Selector(e.X)
	case *ast.CompositeLit:
		return Selector(e.Type)
	case *ast.SelectorExpr:
		return Ident(e.X), Ident(e.Sel)
	}
	return "", ""
}
