// Package analyzers содержит вручную созданные анализаторы, невходящие в стандартный набор
package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitAnalyzer проверяет, что в main функции нет вызова os.Exit
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check using os.Exit in main function",
	Run:  run,
}

// run проверяет, что в main функции нет вызова os.Exit
// Если в main функции есть вызов os.Exit, то возвращает ошибку
// иначе возвращает nil
//
// Параметры:
//   - pass - анализатор
//
// Возвращаемое значение:
//   - interface{}, error
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkg, ok := sel.X.(*ast.Ident)
				if !ok || pkg.Name != "os" || sel.Sel.Name != "Exit" {
					return true
				}

				pass.Reportf(call.Pos(), "can't use os.Exit in main function")
				return true
			})
			return false
		})
	}
	return nil, nil
}
