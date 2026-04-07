package pathconcat_scheme

func schemeConcat(host string) string {
	return "https://" + host // want `use url\.JoinPath\(\) instead of string concatenation with "/"`
}

func schemeConcatHTTP(host string) string {
	return "http://" + host // want `use url\.JoinPath\(\) instead of string concatenation with "/"`
}

func schemeWithSlashSep(base, path string) string {
	return "https://example.com" + "/" + path // want `use url\.JoinPath\(\) instead of string concatenation with "/"`
}
