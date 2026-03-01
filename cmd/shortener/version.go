package main

import (
	"fmt"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	formatValue := func(s string) string {
		if s == "" {
			return "N/A"
		}
		return s
	}

	fmt.Printf("Build version: %s\n", formatValue(buildVersion))
	fmt.Printf("Build date: %s\n", formatValue(buildDate))
	fmt.Printf("Build commit: %s\n", formatValue(buildCommit))
}
