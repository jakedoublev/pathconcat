package pathconcat_context

import (
	"fmt"
	"net/http"
	"os"
)

// --- Flagged: direct argument to known consumer ---

func directOsOpen(dir, file string) {
	os.Open(dir + "/" + file) // want `use filepath\.Join\(\) instead of string concatenation with "/"`
}

func directHTTPGet(base, path string) {
	http.Get(base + "/" + path) // want `use url\.JoinPath\(\) instead of string concatenation with "/"`
}

// --- Flagged: one-hop variable ---

func oneHopOsOpen(dir, file string) {
	p := dir + "/" + file // want `use filepath\.Join\(\) instead of string concatenation with "/"`
	os.Open(p)
}

func oneHopHTTPGet(base, path string) {
	u := base + "/" + path // want `use url\.JoinPath\(\) instead of string concatenation with "/"`
	http.Get(u)
}

// --- NOT flagged: no path consumer context ---

func displayString(a, b string) {
	fmt.Println(a + "/" + b) // OK: fmt.Println is not a path consumer
}

func returned(a, b string) string {
	return a + "/" + b // OK: returned, no consumer context
}

func assignedOnly(a, b string) {
	_ = a + "/" + b // OK: unused
}
