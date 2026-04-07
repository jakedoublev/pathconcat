package pathconcat

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func newAnalyzer(settings Settings) *analysis.Analyzer {
	r := &runner{settings: settings}
	return &analysis.Analyzer{
		Name:     "pathconcat",
		Doc:      "Detects string-based path/URL concatenation; suggests filepath.Join, path.Join, or url.JoinPath",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      r.run,
	}
}

type runner struct {
	settings Settings
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BinaryExpr)(nil),
		(*ast.CallExpr)(nil),
	}

	// Track binary expressions we've already reported to avoid duplicates
	// when we encounter sub-expressions of a chain we already handled.
	reported := map[ast.Node]bool{}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch expr := n.(type) {
		case *ast.BinaryExpr:
			if reported[expr] {
				return
			}
			r.checkBinaryConcat(pass, expr, reported)
		case *ast.CallExpr:
			r.checkSprintfCall(pass, expr)
			r.checkStringsJoin(pass, expr)
		}
	})

	return nil, nil
}

// checkBinaryConcat detects x + "/" + y patterns.
func (r *runner) checkBinaryConcat(pass *analysis.Pass, expr *ast.BinaryExpr, reported map[ast.Node]bool) {
	if expr.Op != token.ADD {
		return
	}

	// Collect all literals in the concatenation chain.
	chain := flattenAddChain(expr)

	// Check if any element in the chain is the literal "/".
	hasSlashSep := false
	for _, node := range chain {
		if lit, ok := node.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			val, err := strconv.Unquote(lit.Value)
			if err != nil {
				continue
			}
			if val == "/" {
				hasSlashSep = true
				break
			}
		}
	}
	if !hasSlashSep {
		return
	}

	// Suppress: chains containing ignored string literals.
	if r.containsIgnoredString(chain) {
		return
	}

	// Suppress: scheme prefix like "https://" + host without further "/" concat.
	if r.isSchemePrefix(chain) {
		return
	}

	// Mark all sub-expressions as reported.
	markChain(expr, reported)

	suggestion := suggestFunc(pass, expr)
	pass.Report(analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "use " + suggestion + "() instead of string concatenation with \"/\"",
	})
}

// checkSprintfCall detects fmt.Sprintf("%s/%s", ...) patterns.
func (r *runner) checkSprintfCall(pass *analysis.Pass, call *ast.CallExpr) {
	if !isFuncCall(pass, call, "fmt", "Sprintf") {
		return
	}

	if len(call.Args) < 2 {
		return
	}

	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}

	format, err := strconv.Unquote(lit.Value)
	if err != nil {
		return
	}

	if !hasPathSeparatorInFormat(format) {
		return
	}

	// Suppress: format strings containing ignored substrings.
	for _, seg := range r.settings.IgnoreStrings {
		if strings.Contains(format, seg) {
			return
		}
	}

	// Suppress: connection strings (postgres://, mysql://, amqp://, etc.)
	if connectionStringPattern.MatchString(format) {
		return
	}

	suggestion := suggestFunc(pass, call)
	pass.Report(analysis.Diagnostic{
		Pos:     call.Pos(),
		End:     call.End(),
		Message: "use " + suggestion + "() instead of fmt.Sprintf with path separators",
	})
}

// checkStringsJoin detects strings.Join(parts, "/").
func (r *runner) checkStringsJoin(pass *analysis.Pass, call *ast.CallExpr) {
	if !isFuncCall(pass, call, "strings", "Join") {
		return
	}

	if len(call.Args) < 2 {
		return
	}

	lit, ok := call.Args[1].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}

	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		return
	}

	if val != "/" {
		return
	}

	suggestion := suggestFunc(pass, call)
	pass.Report(analysis.Diagnostic{
		Pos:     call.Pos(),
		End:     call.End(),
		Message: "use " + suggestion + "() instead of strings.Join with \"/\"",
	})
}

// flattenAddChain collects all operands in a chain of + operations.
func flattenAddChain(expr ast.Expr) []ast.Expr {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.ADD {
		return []ast.Expr{expr}
	}
	return append(flattenAddChain(bin.X), flattenAddChain(bin.Y)...)
}

// markChain marks all BinaryExpr nodes in a + chain as reported.
func markChain(expr ast.Expr, reported map[ast.Node]bool) {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.ADD {
		return
	}
	reported[bin] = true
	markChain(bin.X, reported)
	markChain(bin.Y, reported)
}

// containsIgnoredString returns true if any string literal in the chain
// contains a configured ignore-strings substring.
func (r *runner) containsIgnoredString(chain []ast.Expr) bool {
	for _, node := range chain {
		lit, ok := node.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		val, err := strconv.Unquote(lit.Value)
		if err != nil {
			continue
		}
		for _, seg := range r.settings.IgnoreStrings {
			if strings.Contains(val, seg) {
				return true
			}
		}
	}
	return false
}

// isSchemePrefix returns true if the chain is "https://" + host or
// "http://" + host without a subsequent "/" separator.
func (r *runner) isSchemePrefix(chain []ast.Expr) bool {
	if len(chain) != 2 {
		return false
	}
	lit, ok := chain[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		return false
	}
	return strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://")
}

// sprintfPathPattern matches format strings like "%s/%s" or "%v/%s".
var sprintfPathPattern = regexp.MustCompile(`%[svdqxX]\s*/\s*%[svdqxX]`)

// connectionStringPattern matches DSN/connection string prefixes.
var connectionStringPattern = regexp.MustCompile(`^(postgres|mysql|amqp|redis|mongodb|nats)://`)

// hasPathSeparatorInFormat returns true if a format string contains
// a pattern like %s/%s indicating path construction.
func hasPathSeparatorInFormat(format string) bool {
	return sprintfPathPattern.MatchString(format)
}

// isFuncCall returns true if call resolves to pkg.name.
func isFuncCall(pass *analysis.Pass, call *ast.CallExpr, pkg, name string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}

	obj := pass.TypesInfo.ObjectOf(sel.Sel)
	if obj == nil {
		return false
	}

	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}

	return fn.Pkg() != nil && fn.Pkg().Path() == pkg
}

// suggestFunc determines the best join function based on file imports and context.
func suggestFunc(pass *analysis.Pass, node ast.Node) string {
	file := fileForNode(pass, node)
	if file == nil {
		return "path.Join"
	}

	hasImport := func(path string) bool {
		for _, imp := range file.Imports {
			impPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if impPath == path {
				return true
			}
		}
		return false
	}

	// Check if context suggests URL construction.
	if hasImport("net/url") || hasImport("net/http") {
		return "url.JoinPath"
	}

	// Check for string literals starting with http(s)://.
	if containsHTTPScheme(node) {
		return "url.JoinPath"
	}

	// Check if context suggests filesystem paths.
	if hasImport("path/filepath") || hasImport("os") {
		return "filepath.Join"
	}

	return "path.Join"
}

// fileForNode finds the *ast.File containing the given node.
func fileForNode(pass *analysis.Pass, node ast.Node) *ast.File {
	pos := node.Pos()
	for _, file := range pass.Files {
		if file.Pos() <= pos && pos < file.End() {
			return file
		}
	}
	return nil
}

// containsHTTPScheme checks if a node's concatenation chain contains
// a string literal starting with http:// or https://.
func containsHTTPScheme(node ast.Node) bool {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	chain := flattenAddChain(bin)
	for _, n := range chain {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		val, err := strconv.Unquote(lit.Value)
		if err != nil {
			continue
		}
		if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") {
			return true
		}
	}
	return false
}
