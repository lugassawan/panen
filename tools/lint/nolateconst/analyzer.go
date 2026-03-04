// Package nolateconst defines an analyzer that forbids package-level const
// and var declarations that appear after any function or method declaration.
// All constants and variables should be grouped before functions.
package nolateconst

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports package-level const/var declarations that appear after function declarations.
var Analyzer = &analysis.Analyzer{
	Name: "nolateconst",
	Doc:  "forbids package-level const/var declarations after function declarations",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		seenFunc := false
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				seenFunc = true
			case *ast.GenDecl:
				if !seenFunc {
					continue
				}
				switch d.Tok {
				case token.CONST:
					pass.Reportf(d.TokPos, "const declaration after function declaration")
				case token.VAR:
					pass.Reportf(d.TokPos, "var declaration after function declaration")
				}
			}
		}
	}
	return nil, nil
}
