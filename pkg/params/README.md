# Params Package

The **params** package defines helper types for parsing pagination, ordering, and filtering parameters commonly used in HTTP APIs.

## Usage

- Paginate results: `?limit=10&offset=20` => `&Page{Offset: 20, Limit: 10}`
- Order by field: `?orderBy=-priority,created_at` => `[]*OrderBy{&OrderBy{Col: "priority", Dir: true}, &OrderBy{Col: "created_at"}}`
- Apply filters: `?strings=active` => `&Filter{Strings: []string{"active"}}`

## Examples

```go
import (
    "github.com/nojyerac/go-lib/pkg/params"
)

// define a custom filter struct using tags for names & validation
type Filter struct {
    Strings   []string      `params:"strings,modify"`
    Ints      []int         `params:"ints"`
    StrPtr    *string       `params:"strPtr"`
    Dur       time.Duration `params:"duration"`
    Time      time.Time     `params:"time"`
    Omit      int           `params:"-"`
    Unmarsher []*IntFilter  `params:"tu"`
    NoTag1    uint          `validate:"lt=10"`
}

// use the struct in a handler
func Handler(c *gin.Context) {
    f := &Filter{}
    if err := params.GetFilter(f, c); err != nil {
        // handle error
    }
    p, err := params.GetPage(c)
    if err != nil {
        // handle err
    }
    o, err := params.GetOrderBy(c)
    if err != nil {
        // handle err
    }
}
```
