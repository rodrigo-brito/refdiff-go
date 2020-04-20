package main

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func GetCallGraph(dir string, file string) (map[string][]string, error) {
	var (
		defaultPackage  string
		firstFolderName string
	)

	parts := strings.Split(file, "/")
	if len(parts) > 1 {
		defaultPackage = strings.Join(parts[:len(parts)-1], "/")
		firstFolderName = parts[0]
	}

	calls := make(map[string][]string)
	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: false,
		Dir:   dir,
	}
	initial, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(initial) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	// Create and build SSA-form program representation.
	prog, _ := ssautil.AllPackages(initial, 0)
	prog.Build()

	var cg = static.CallGraph(prog)
	cg.DeleteSyntheticNodes()

	if err := callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		if !isValid(edge) {
			return nil
		}

		caller := formatFunctionName(edge.Caller.Func, defaultPackage)
		callee := formatFunctionName(edge.Callee.Func, defaultPackage)
		if strings.HasPrefix(callee, firstFolderName) {
			calls[caller] = append(calls[caller], callee)
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return calls, nil
}

func isValid(edge *callgraph.Edge) bool {
	caller := edge.Caller.Func
	callee := edge.Caller.Func

	return caller.Pkg != nil &&
		caller.Pkg.Pkg != nil &&
		callee.Name() != "init"
}

func trimReceiver(name string) string {
	name = strings.ReplaceAll(name, "*command-line-arguments.", "")
	name = strings.ReplaceAll(name, "command-line-arguments.", "")
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func normalizePackage(pkg, defaultPackage string) string {
	if pkg == "command-line-arguments" {
		return defaultPackage
	}

	firstFolder := strings.SplitN(defaultPackage, "/", 2)[0]
	index := strings.Index(pkg, firstFolder)
	if index >= 0 {
		return pkg[index:]
	}

	return pkg
}

func formatFunctionName(f *ssa.Function, defaultPackage string) string {
	singature := ""
	if f.Pkg != nil && f.Pkg.Pkg != nil {
		singature = normalizePackage(f.Pkg.Pkg.Path(), defaultPackage) + "/"
	}

	if f.Signature.Recv() != nil && f.Signature.Recv().Type().String() != "" {
		singature += trimReceiver(f.Signature.Recv().Type().String()) + "."
	}

	return fmt.Sprintf("%s%s", singature, f.Name())
}
