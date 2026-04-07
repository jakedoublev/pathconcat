package pathconcat

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("pathconcat", New)
}

// Settings configures the pathconcat linter.
type Settings struct {
	// IgnoreStrings is a list of substrings that suppress diagnostics.
	// If any string literal in a concatenation chain or fmt.Sprintf format
	// string contains one of these substrings, the finding is skipped.
	// Useful for domain-specific identifiers built with "/" that are not
	// file or URL paths (e.g. "/attr/", "/value/", "::").
	IgnoreStrings []string `json:"ignore-strings"`
}

// Plugin implements register.LinterPlugin.
type Plugin struct {
	settings Settings
}

// New creates a new pathconcat plugin from golangci-lint settings.
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	return &Plugin{settings: s}, nil
}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	a := NewAnalyzer(p.settings)
	return []*analysis.Analyzer{a}, nil
}

func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
