package params

import (
	"strconv"

	"github.com/gin-gonic/gin"
	apierrors "github.com/go-openapi/errors"
)

// Page is the data structure for pagination params
type Page struct {
	Limit  int64
	Offset int64
}

// GetPage returns the pagination params from a request
func GetPage(c *gin.Context) (*Page, error) {
	page := &Page{}
	if o := c.Query(offsetParam); o != "" {
		offset, err := strconv.ParseInt(o, 10, 64)
		if err != nil {
			return nil, apierrors.NewParseError(offsetParam, query, o, err)
		}
		page.Offset = offset
	}
	if l := c.Query(limitParam); l != "" {
		limit, err := strconv.ParseInt(l, 10, 64)
		if err != nil {
			return nil, apierrors.NewParseError(limitParam, query, l, err)
		}
		page.Limit = limit
	}
	return page, nil
}
