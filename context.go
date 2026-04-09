package pathconcat

import (
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

// pathConsumers maps package paths to function names that consume file paths or URLs.
var pathConsumers = map[string]map[string]bool{
	"os": {
		"Open": true, "Create": true, "OpenFile": true,
		"Stat": true, "Lstat": true,
		"Mkdir": true, "MkdirAll": true,
		"Remove": true, "RemoveAll": true, "Rename": true,
		"ReadFile": true, "WriteFile": true, "ReadDir": true,
		"Chdir": true, "Chmod": true, "Chown": true,
	},
	"path/filepath": {
		"Join": true, "Abs": true, "Dir": true, "Base": true,
		"Clean": true, "Ext": true, "Glob": true,
		"Walk": true, "WalkDir": true, "Rel": true, "Match": true,
		"EvalSymlinks": true,
	},
	"net/http": {
		"Get": true, "Post": true, "Head": true,
		"NewRequest": true, "NewRequestWithContext": true,
	},
	"net/url": {
		"Parse": true, "JoinPath": true,
	},
}

// consumerKind classifies which domain a consumer function belongs to.
type consumerKind int

const (
	consumerNone consumerKind = iota
	consumerFilepath
	consumerURL
)

// checkPathContext uses SSA to determine whether the concatenation spanning
// [start, end) flows into a known path/URL-consuming function. It finds the
// outermost SSA BinOp within the expression range and follows its referrers.
func checkPathContext(pass *analysis.Pass, start, end token.Pos) consumerKind {
	ssaResult, ok := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	if !ok {
		return consumerNone
	}

	// Find SSA BinOp instructions whose position falls within the AST range.
	// The outermost one (latest position in the chain) is the final result.
	var candidate ssa.Value
	for _, fn := range ssaResult.SrcFuncs {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				instrPos := instr.Pos()
				if instrPos < start || instrPos >= end {
					continue
				}
				binOp, ok := instr.(*ssa.BinOp)
				if !ok {
					continue
				}
				// Keep the last BinOp in the range — it produces the final concatenated value.
				candidate = binOp
			}
		}
	}

	if candidate == nil {
		return consumerNone
	}
	return walkReferrers(candidate, 0)
}

// walkReferrers follows the referrer chain of a value to find if it reaches
// a known path/URL consumer. maxDepth limits recursion to prevent cycles.
func walkReferrers(val ssa.Value, depth int) consumerKind {
	if depth > 3 {
		return consumerNone
	}

	refs := val.Referrers()
	if refs == nil {
		return consumerNone
	}

	for _, ref := range *refs {
		switch instr := ref.(type) {
		case *ssa.Call:
			if kind := classifySSACall(instr); kind != consumerNone {
				return kind
			}
		case *ssa.Phi:
			// Value merged from control flow — follow through.
			if kind := walkReferrers(instr, depth+1); kind != consumerNone {
				return kind
			}
		case *ssa.Store:
			// Stored to address — follow the address's referrers.
			if kind := walkReferrers(instr.Addr, depth+1); kind != consumerNone {
				return kind
			}
		case ssa.Value:
			// Other value-producing instructions — follow referrers.
			if kind := walkReferrers(instr, depth+1); kind != consumerNone {
				return kind
			}
		}
	}
	return consumerNone
}

// classifySSACall checks if a call instruction targets a known path/URL consumer.
func classifySSACall(call *ssa.Call) consumerKind {
	callee := call.Call.StaticCallee()
	if callee == nil {
		return consumerNone
	}

	pkg := callee.Package()
	if pkg == nil {
		return consumerNone
	}

	pkgPath := pkg.Pkg.Path()
	funcs, ok := pathConsumers[pkgPath]
	if !ok {
		return consumerNone
	}

	if !funcs[callee.Name()] {
		return consumerNone
	}

	switch pkgPath {
	case "os", "path/filepath":
		return consumerFilepath
	case "net/http", "net/url":
		return consumerURL
	}
	return consumerNone
}

// suggestFromContext returns a suggestion based on the consumer kind.
func suggestFromContext(kind consumerKind) string {
	switch kind {
	case consumerFilepath:
		return "filepath.Join"
	case consumerURL:
		return "url.JoinPath"
	default:
		return "path.Join"
	}
}
