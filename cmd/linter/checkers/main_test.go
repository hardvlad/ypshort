package checkers

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestSafeCheck(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "panic_in_main",
			files: map[string]string{
				"a/main.go": `
package main

import (
	"log"
)

func main() {
	panic("test") // want "использование panic запрещено"
	log.Fatal("fatal in main — OK")
}
`,
			},
		},
		{
			name: "panic_outside_main",
			files: map[string]string{
				"a/main.go": `
package main

func helper() {
	panic("bad") // want "использование panic запрещено"
}

func main() {}
`,
			},
		},
		{
			name: "os.Exit_outside_main",
			files: map[string]string{
				"a/main.go": `
package main

import (
	"os"
)

func init() {
	os.Exit(1) // want "вызов os.Exit возможен только в функции main()"
}

func main() {}
`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewAnalyzer()
			dir, cleanup, err := analysistest.WriteFiles(test.files)
			if err != nil {
				t.Fatal(err)
			}

			analysistest.Run(t, dir, a, "a")

			cleanup()
		})
	}
}
