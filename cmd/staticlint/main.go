package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"

	custom "github.com/FollowLille/metrics/cmd/staticlint/analyzer"
)

func main() {
	var mychecks []*analysis.Analyzer

	// Добавляем стандартные анализаторы
	mychecks = append(mychecks,
		shadow.Analyzer,
		structtag.Analyzer,
		nilness.Analyzer,
		custom.NoOsExitAnalyzer,
	)

	checks := staticcheck.Analyzers
	for _, a := range checks {
		if a.Analyzer.Name == "SA" || a.Analyzer.Name == "ST" {
			mychecks = append(mychecks, a.Analyzer)
		}
	}

	multichecker.Main(mychecks...)
}
