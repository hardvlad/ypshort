package checkers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "safecheck",
		Doc:  "проверяет использование panic и небезопасных выходов (log.Fatal/os.Exit вне main)",
		Run:  run,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				if id, ok := getFuncName(pass, node.Fun); ok {
					handleCall(pass, node, id)
				}
			}
			return true
		})
	}
	return nil, nil
}

// getFuncName извлекает имя функции (для вызова вида panic(...), log.Fatal(...), os.Exit(...))
func getFuncName(pass *analysis.Pass, fn ast.Expr) (string, bool) {
	switch expr := fn.(type) {
	case *ast.Ident:
		return expr.Name, true

	case *ast.SelectorExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			obj, ok := pass.TypesInfo.Uses[ident]
			if !ok {
				return "", false
			}
			if pkg := obj.Pkg(); pkg != nil {
				return pkg.Name() + "." + expr.Sel.Name, true
			}
			return "", false
		}
	}
	return "", false
}

// handleCall проверяет вызовы panic, log.Fatal и os.Exit
func handleCall(pass *analysis.Pass, call *ast.CallExpr, funcID string) {
	switch funcID {
	case "panic":
		pass.Report(analysis.Diagnostic{
			Pos:     call.Pos(),
			Message: "использование panic запрещено",
		})

	case "log.Fatal", "log.Fatalf", "log.Fatalln":
		if !inMainFunc(pass, call) {
			pass.Report(analysis.Diagnostic{
				Pos:     call.Pos(),
				Message: "вызов log.Fatal возможен только в функции main() пакета main",
			})
		}

	case "os.Exit":
		if !inMainFunc(pass, call) {
			pass.Report(analysis.Diagnostic{
				Pos:     call.Pos(),
				Message: "вызов os.Exit возможен только в функции main() пакета main",
			})
		}
	}
}

// inMainFunc проверяет, находится ли вызов внутри функции func main()
func inMainFunc(pass *analysis.Pass, node ast.Node) bool {
	for _, file := range pass.Files {
		if !isInMainPackage(pass) {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Recv != nil {
				continue
			}

			if containsNode(fn.Body, node) {
				return true
			}
		}
	}
	return false
}

// isInMainPackage — проверка: пакет называется main
func isInMainPackage(pass *analysis.Pass) bool {
	return pass.Pkg.Name() == "main"
}

// containsNode проверяет, находится ли child внутри parent в AST (включая глубину)
func containsNode(parent, child ast.Node) bool {
	found := false
	ast.Inspect(parent, func(n ast.Node) bool {
		if n == child {
			found = true
			return false
		}
		return true
	})
	return found
}
