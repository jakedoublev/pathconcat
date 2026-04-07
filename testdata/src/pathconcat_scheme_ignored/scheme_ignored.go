package pathconcat_scheme_ignored

func schemeWithIgnoredString(ns, name string) string {
	return "https://namespace.com" + "/attr/" + name // OK: ignore-strings takes priority over check-scheme-concat
}

func schemeAlone(host string) string {
	return "https://" + host // want `use url\.JoinPath\(\) instead of string concatenation with "/"`
}
