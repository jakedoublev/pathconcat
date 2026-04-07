package pathconcat

import (
	"fmt"
	"strings"
)

// --- True positives: should be flagged ---

func concatWithSlash(a, b string) string {
	return a + "/" + b // want `use path\.Join\(\) instead of string concatenation with "/"`
}

func concatMultiSegment(a, b, c string) string {
	return a + "/" + b + "/" + c // want `use path\.Join\(\) instead of string concatenation with "/"`
}

func sprintfPath(a, b string) string {
	return fmt.Sprintf("%s/%s", a, b) // want `use path\.Join\(\) instead of fmt\.Sprintf with path separators`
}

func sprintfMultiSegment(a, b, c string) string {
	return fmt.Sprintf("%s/%s/%s", a, b, c) // want `use path\.Join\(\) instead of fmt\.Sprintf with path separators`
}

func stringsJoinSlash(parts []string) string {
	return strings.Join(parts, "/") // want `use path\.Join\(\) instead of strings\.Join with "/"`
}

// --- False positives: should NOT be flagged ---

func schemePrefix(host string) string {
	return "https://" + host // Scheme prefix, only 2 elements
}

func regularConcat(a, b string) string {
	return a + b // No slash separator
}

func concatNonSep(a string) string {
	return a + "/api" // No bare "/" separator element
}

func sprintfNoPathSep(a string) string {
	return fmt.Sprintf("value: %s", a) // No path separator in format
}

func stringsJoinComma(parts []string) string {
	return strings.Join(parts, ",") // Not a slash separator
}

func postgresConnStr(user, pass, host, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pass, host, db) // connection string
}

func mysqlConnStr(user, pass, host, db string) string {
	return fmt.Sprintf("mysql://%s:%s@%s/%s", user, pass, host, db) // connection string
}

func redisConnStr(host string, port, db int) string {
	return fmt.Sprintf("redis://%s:%d/%d", host, port, db) // connection string
}

func mongodbConnStr(host, db string) string {
	return fmt.Sprintf("mongodb://%s/%s", host, db) // connection string
}

func amqpConnStr(user, pass, host, vhost string) string {
	return fmt.Sprintf("amqp://%s:%s@%s/%s", user, pass, host, vhost) // connection string
}

func natsConnStr(host, subject string) string {
	return fmt.Sprintf("nats://%s/%s", host, subject) // connection string
}

func sprintfMixedVerbs(a string, b int) string {
	return fmt.Sprintf("%s/%d", a, b) // want `use path\.Join\(\) instead of fmt\.Sprintf with path separators`
}

func sprintfWithStaticPrefix(base, name string) string {
	return fmt.Sprintf("/api/%s/%s", base, name) // want `use path\.Join\(\) instead of fmt\.Sprintf with path separators`
}

func httpSprintfPath(host, path string) string {
	return fmt.Sprintf("http://%s/%s", host, path) // want `use path\.Join\(\) instead of fmt\.Sprintf with path separators`
}
