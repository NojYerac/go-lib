# Version Package

The `version` package provides runtime and build metadata.

## API

- `GetVersion() Version`
- `SetServiceName(name string)`
- `SetSemVer(ver string)`

```go
type Version struct {
    Name   string `json:"serviceName"`
    SemVer string `json:"semVer"`
    GitSHA string `json:"gitSHA,omitempty"`
    GoVer  string `json:"goVer"`
    Arch   string `json:"arch"`
    OS     string `json:"os"`
}
```

## Defaults

- `SemVer`: `0.0.0`
- `Name`: empty string
- `GitSHA`: empty string

`GoVer`, `Arch`, and `OS` are always populated from `runtime`.

## Example

```go
version.SetServiceName("orders")
version.SetSemVer("1.4.0")

v := version.GetVersion()
fmt.Printf("%s %s (%s)\n", v.Name, v.SemVer, v.GitSHA)
```

## Build-Time Git SHA

You can inject `gitSHA` at build time:

```bash
go build -ldflags "-X github.com/nojyerac/go-lib/version.gitSHA=$(git rev-parse --short HEAD)"
```
