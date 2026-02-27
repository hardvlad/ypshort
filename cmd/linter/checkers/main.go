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
				if id, ok := getFuncName(node.Fun); ok {
					handleCall(pass, node, id)
				}
			}
			return true
		})
	}
	return nil, nil
}

// getFuncName извлекает имя функции (для вызова вида panic(...), log.Fatal(...), os.Exit(...))
func getFuncName(fn ast.Expr) (string, bool) {
	switch expr := fn.(type) {
	case *ast.Ident:
		return expr.Name, true
	case *ast.SelectorExpr:
		if x, ok := expr.X.(*ast.Ident); ok && x.Name != "" {
			return x.Name + "." + expr.Sel.Name, true
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
	var currentFunc *ast.FuncDecl

	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" &&
			isInMainPackage(pass) {

			// Ищем, входит ли `node` в тело этой функции
			for _, stmt := range fn.Body.List {
				if containsNode(stmt, node) {
					currentFunc = fn
					return false
				}
			}
		}
		return true
	})

	return currentFunc != nil
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
