# pathconcat

A [golangci-lint](https://golangci-lint.run) module plugin that detects string-based path and URL concatenation in Go code, suggesting `path.Join`, `filepath.Join`, or `url.JoinPath` instead.

## What it catches

| Pattern | Example | Suggestion |
|---------|---------|------------|
| String concat with `"/"` | `a + "/" + b` | `path.Join(a, b)` |
| `fmt.Sprintf` with path separators | `fmt.Sprintf("%s/%s", a, b)` | `path.Join(a, b)` |
| `strings.Join` with `"/"` | `strings.Join(parts, "/")` | `path.Join(parts...)` |

### Smart suggestions

The suggested function depends on file context:

- **`url.JoinPath`** if the file imports `net/url` or `net/http`, or the concatenation contains `http://`/`https://`
- **`filepath.Join`** if the file imports `path/filepath` or `os`
- **`path.Join`** otherwise

### Built-in suppression

- **Connection strings**: `fmt.Sprintf("postgres://%s:%s@%s/%s", ...)` is not flagged
- **Scheme prefixes**: `"https://" + host` is not flagged (unless `check-scheme-concat` is enabled)
- **Ignored strings**: Configurable substrings (e.g. `/attr/`, `/value/`) that indicate domain-specific identifier construction rather than path building

## Installation

Requires [golangci-lint](https://golangci-lint.run) v2 with the [module plugin system](https://golangci-lint.run/docs/plugins/module-plugins/).

### 1. Create `.custom-gcl.yml`

```yaml
version: v2.8.0
plugins:
  - module: github.com/jakedoublev/pathconcat
    version: v0.3.0
```

### 2. Add to `.golangci.yaml`

```yaml
linters:
  enable:
    - pathconcat
  settings:
    custom:
      pathconcat:
        type: module
        description: "Detects string path/URL concatenation"
        settings:
          # Optional: suppress findings where any literal contains these substrings
          ignore-strings:
            - "/attr/"
            - "/value/"
          # Optional: flag "https://" + host scheme concatenation
          # check-scheme-concat: true
```

### 3. Build and run

```bash
golangci-lint custom     # builds ./custom-gcl with the plugin
./custom-gcl run ./...   # run linting
```

## Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `ignore-strings` | `[]string` | `[]` | Substrings that suppress diagnostics. If any string literal in a concatenation chain or `fmt.Sprintf` format string contains one of these, the finding is skipped. |
| `check-scheme-concat` | `bool` | `false` | When true, also flags scheme prefix concatenation like `"https://" + host`. By default these are not flagged. |

## Suppressing individual findings

Use the standard `//nolint:pathconcat` directive:

```go
resource := p[1] + "/" + p[2] //nolint:pathconcat // reconstructing gRPC procedure path
```

## License

MIT
