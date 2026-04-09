package pathconcat_backslash

import (
	"fmt"
	"strings"
)

// --- Backslash separator (Windows-style paths) ---

func backslashConcat(a, b string) string {
	return a + "\\" + b // want `use path\.Join\(\) instead of string concatenation with "/"`
}

func backslashMultiSegment(a, b, c string) string {
	return a + "\\" + b + "\\" + c // want `use path\.Join\(\) instead of string concatenation with "/"`
}

func stringsJoinBackslash(parts []string) string {
	return strings.Join(parts, "\\") // want `use path\.Join\(\) instead of strings\.Join with "/"`
}

// --- Mixed separators (cross-platform bugs) ---

func mixedSeparators(a, b, c string) string {
	return a + "/" + b + "\\" + c // want `use path\.Join\(\) instead of string concatenation with "/"`
}

// --- Windows drive path construction ---

func drivePathConcat(drive, dir, file string) string {
	return drive + "\\" + dir + "\\" + file // want `use path\.Join\(\) instead of string concatenation with "/"`
}

// --- Sprintf with backslash ---

func sprintfBackslash(a, b string) string {
	return fmt.Sprintf("%s\\%s", a, b) // OK: backslash in sprintf is not detected (not a path separator pattern)
}

// --- False positives: should NOT be flagged ---

func escapeSequence(a string) string {
	return a + "\n" // OK: newline, not a path separator
}

func regularBackslashInString(a string) string {
	return a + "\\n" // OK: not a bare "\\" separator
}
