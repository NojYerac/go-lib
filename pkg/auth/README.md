# Auth Package

The **auth** package provides simple authentication helpers and user data structures.

## Configuration

This package does not expose any configuration structs. All logic operates on the `User` type and the error type `ErrUnauthenticated`.

## Usage

```go
package main

import (
    "github.com/nojyerac/go-lib/pkg/auth"
)

func main() {
    // Create a user
    user := &auth.User{UserID: 1, Username: "alice"}
    // Handle authentication logic elsewhere
    _ = user
}
```

## Examples

- **Creating a User** – Construct a `User` with privileges and features.
- **Error Handling** – Return `auth.ErrUnauthenticated` when a request lacks credentials.
