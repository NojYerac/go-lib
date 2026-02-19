# Version Package

The **version** package stores build‑time metadata such as the service name, semantic version, Git SHA, Go runtime version, architecture and OS.

## Configuration

The package does not expose configuration structs; it simply provides helper functions.

### Accessing Version Information

```go
import "github.com/nojyerac/go-lib/pkg/version"

func main() {
    v := version.GetVersion()
    fmt.Printf("%s %s (%s)\n", v.Name, v.SemVer, v.GitSHA)
}
```

The values can be injected at build time via `-ldflags`:

```bash
go build -ldflags "-X version.gitSHA=$(git rev-parse --short HEAD)"
```

## Usage

```go
import (
    "github.com/nojyerac/go-lib/pkg/version"
    "log"
)

func main() {
    v := version.GetVersion()
    log.Printf("%s %s", v.Name, v.SemVer)
}
```

## Examples

- Display the version string in a CLI application.
- Expose a `/version` endpoint in an HTTP server.
- Use the Git SHA for debugging builds.
