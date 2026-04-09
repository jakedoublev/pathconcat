package pathconcat

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestPathConcat(t *testing.T) {
	newPlugin, err := register.GetPlugin("pathconcat")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[0], "pathconcat")
}

func TestCheckSchemeConcat(t *testing.T) {
	a := NewAnalyzer(Settings{CheckSchemeConcat: true})

	analysistest.Run(t, testdataDir(t), a, "pathconcat_scheme")
}

func TestCheckSchemeConcatWithIgnoreStrings(t *testing.T) {
	a := NewAnalyzer(Settings{
		CheckSchemeConcat: true,
		IgnoreStrings:     []string{"/attr/"},
	})

	analysistest.Run(t, testdataDir(t), a, "pathconcat_scheme_ignored")
}

func TestRequirePathContext(t *testing.T) {
	a := NewAnalyzer(Settings{RequirePathContext: true})

	analysistest.Run(t, testdataDir(t), a, "pathconcat_context")
}

func TestBackslashConcat(t *testing.T) {
	a := NewAnalyzer(Settings{})

	analysistest.Run(t, testdataDir(t), a, "pathconcat_backslash")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		require.Fail(t, "unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
