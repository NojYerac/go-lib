package params

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	apierrors "github.com/go-openapi/errors"
)

// OrderBy is the data structure for order by params
type OrderBy struct {
	Column string
	Dir    Dir
}

// String implements the fmt.Stringer interface
func (o OrderBy) String() string {
	return fmt.Sprintf("%s %s", o.Column, o.Dir)
}

// GetOrderBy returns the order by params from a request
func GetOrderBy(c *gin.Context) ([]*OrderBy, error) {
	var orderBys []*OrderBy
	if o := c.Query(orderByParam); o != "" {
		for _, ob := range strings.Split(o, ",") {
			if !orderByRE.MatchString(ob) {
				return nil, apierrors.NewParseError(orderByParam, query, ob, errInvalidFormat)
			}
			orderBy := &OrderBy{
				Column: ob,
			}
			if ob[0] == '-' {
				orderBy.Dir = DESC
				orderBy.Column = ob[1:]
			}
			orderBys = append(orderBys, orderBy)
		}
	}
	return orderBys, nil
}

var (
	orderByRE        = regexp.MustCompile(`^-?[a-zA-Z0-9_-]+$`)
	errInvalidFormat = errors.New("invalid format")
)

// Dir is the direction data type for order by params
type Dir bool

// String implements the fmt.Stringer interface
func (d Dir) String() string {
	if d {
		return "DESC"
	}
	return "ASC"
}

const (
	// ASC ascending
	ASC Dir = false
	// DESC descending
	DESC Dir = true
)
