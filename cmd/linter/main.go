package main

import (
	"github.com/hardvlad/ypshort/linter/checkers"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(checkers.NewAnalyzer())
}
