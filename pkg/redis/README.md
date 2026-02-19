# Redis Package

The **redis** package offers a simple wrapper around a Redis client, exposing connection options and basic utilities.

## Configuration

```go
// pkg/redis/config.go
package redis

import "time"

// Configuration holds Redis connection settings.
//
//   Address   string `config:"redis_address"`
//   DB        int    `config:"redis_db"`
//   Password  string `config:"redis_password"`
//   PoolSize  int    `config:"redis_pool_size"`
//   Timeout   time.Duration `config:"redis_timeout"`
//
// The NewConfiguration helper sets defaults such as localhost:6379.
```

### Client

```go
func NewClient(cfg *Configuration) (*redis.Client, error) {
    // Initialise the client using the cfg values.
}
```

## Usage

```go
import (
    "github.com/nojyerac/go-lib/pkg/redis"
    "github.com/go-redis/redis/v8"
)

func main() {
    cfg := redis.NewConfiguration()
    client, err := redis.NewClient(cfg)
    if err != nil {
        log.Fatal().Err(err).Msg("redis init")
    }
    // Use client to get/set keys.
}
```

## Examples

- Connect to a local Redis instance.
- Perform simple GET/SET operations.
- Use connection pooling for concurrent workloads.
