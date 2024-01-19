package params

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	apierrors "github.com/go-openapi/errors"
	"github.com/go-playground/validator/v10"
)

// ErrInvalidParam is returned when param vadildation fails
type ErrInvalidParam struct {
	Name  string
	Value string
}

func (err *ErrInvalidParam) Error() string {
	return fmt.Sprintf("'%s' is invalid for '%s'", err.Value, err.Name)
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

// parameters
const (
	query        = "query"
	orderByParam = "orderBy"
	limitParam   = "limit"
	offsetParam  = "offset"
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

type validater interface {
	Validate() error
}

var v = validator.New()

type setterFunc func(string) (reflect.Value, error)

type paramTag struct {
	Name      string
	Modifiers []string
	Index     int
}

var (
	cache      = map[reflect.Type]map[*paramTag]setterFunc{}
	trueValues = []string{"true", "TRUE", "t", "1"}
)

func parseTag(t reflect.StructTag) (*paramTag, bool) {
	p := &paramTag{}
	tag := t.Get("params")
	if len(tag) < 1 {
		return p, false
	}
	parts := strings.Split(tag, ",")
	p.Name = parts[0]
	if len(parts) > 1 {
		p.Modifiers = parts[1:]
	}
	return p, true
}

func getSetters(rt reflect.Type) map[*paramTag]setterFunc {
	if setters, ok := cache[rt]; ok {
		return setters
	}
	numField := rt.NumField()
	setters := make(map[*paramTag]setterFunc)
	for i := 0; i < numField; i++ {
		field := rt.Field(i)
		tagVal, ok := parseTag(field.Tag)
		if tagVal.Name == "-" {
			continue
		}
		tagVal.Index = i
		if !ok {
			tagVal.Name = strings.ToLower(field.Name)
		}
		setters[tagVal] = getSetter(field.Type, tagVal.Modifiers)
	}
	cache[rt] = setters
	return setters
}

func makePtrSetter(rt reflect.Type, pSetter setterFunc) setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		var pv reflect.Value
		pv, err = pSetter(val)
		if err != nil {
			return
		}
		rv = reflect.New(rt.Elem())
		rv.Elem().Set(pv)
		return
	}
}

func makeSliceSetter(rt reflect.Type, iSetter setterFunc) setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		vals := strings.Split(val, ",")
		rv = reflect.MakeSlice(rt, len(vals), len(vals))
		for i, v := range vals {
			var iVal reflect.Value
			iVal, err = iSetter(v)
			if err != nil {
				return
			}
			rv.Index(i).Set(iVal)
		}
		return
	}
}

func makeBoolSetter() setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		rv = reflect.ValueOf(false)
		for _, t := range trueValues {
			if val == t {
				rv = reflect.ValueOf(true)
				return
			}
		}
		return
	}
}

func makeDurationSetter() setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		var t time.Duration
		t, err = time.ParseDuration(val)
		if err != nil {
			return
		}
		rv = reflect.ValueOf(t)
		return
	}
}

func makeIntSetter(rt reflect.Type) setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		var num int64
		num, err = strconv.ParseInt(val, 10, int(rt.Size()*8))
		if err != nil {
			return
		}
		rv = reflect.New(rt).Elem()
		rv.SetInt(num)
		return
	}
}

func makeUintSetter(rt reflect.Type) setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		var num uint64
		num, err = strconv.ParseUint(val, 10, int(rt.Size()*8))
		if err != nil {
			return
		}
		rv = reflect.New(rt).Elem()
		rv.SetUint(num)
		return
	}
}

func makeTimeSetter() setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		var t time.Time
		t, err = time.Parse(time.RFC3339, val)
		if err != nil {
			return
		}
		rv = reflect.ValueOf(t)
		return
	}
}

func makeUnmarshalerSetter(rt reflect.Type) setterFunc {
	return func(val string) (rv reflect.Value, err error) {
		rv = reflect.New(rt.Elem())
		args := []reflect.Value{
			reflect.ValueOf([]byte(val)),
		}
		returned := rv.MethodByName("UnmarshalText").Call(args)
		if !returned[0].IsNil() {
			err = returned[0].Interface().(error)
		}
		return
	}
}

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func getSetter(rt reflect.Type, mods []string) setterFunc {
	var _ = len(mods) // may need mods to tweak setters

	if rt.Implements(textUnmarshalerType) {
		return makeUnmarshalerSetter(rt)
	}
	switch rt.Kind() {
	case reflect.Ptr:
		pSetter := getSetter(rt.Elem(), mods)
		return makePtrSetter(rt, pSetter)
	case reflect.Slice:
		rte := rt.Elem()
		iSetter := getSetter(rte, mods)
		return makeSliceSetter(rt, iSetter)
	case reflect.Bool:
		return makeBoolSetter()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rt.String() == "time.Duration" {
			return makeDurationSetter()
		}
		return makeIntSetter(rt)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return makeUintSetter(rt)
	case reflect.String:
		return func(val string) (rv reflect.Value, err error) {
			rv = reflect.ValueOf(val)
			return
		}
	case reflect.Struct:
		if rt.String() == "time.Time" {
			return makeTimeSetter()
		}
		fallthrough
	default:
		// TODO: implement TextUnmarshaler detection
		return func(val string) (rv reflect.Value, err error) {
			err = fmt.Errorf("unsupported type %s", rt.String())
			return
		}
	}
}

// GetFilters extracts filters from the query string
func GetFilters(filter interface{}, c *gin.Context) error {
	rv := reflect.ValueOf(filter)
	rt := rv.Type()
	if rt.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer got %s", rt.String())
	}
	rve := rv.Elem()
	rte := rve.Type()
	for paramTag, setter := range getSetters(rte) {
		val := c.Query(paramTag.Name)
		if val != "" {
			fv, err := setter(val)
			if err != nil {
				return apierrors.NewParseError(paramTag.Name, query, val, err)
			}
			rve.Field(paramTag.Index).Set(fv)
		}
	}
	var err error
	if f, ok := filter.(validater); ok {
		err = f.Validate()
	} else {
		err = v.Struct(filter)
	}
	if err != nil {
		return apierrors.New(
			422,
			"failed to validate filter with error: %s; parsing: %+v",
			err.Error(),
			c.Request.URL.RawQuery,
		)
	}
	return nil
}

var _ encoding.TextUnmarshaler = &IntFilter{}
var _ encoding.TextUnmarshaler = &StringFilter{}

type IntFilter struct {
	Not         bool
	Equals      int64
	LessThan    int64
	GreaterThan int64
}

func (i *IntFilter) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	if text[0] == '!' {
		i.Not = true
		text = text[1:]
	}
	if text[0] == '<' {
		lt, err := strconv.ParseInt(string(text[1:]), 10, 64)
		if err != nil {
			return err
		}
		i.LessThan = lt
		return nil
	}
	if text[0] == '>' {
		gt, err := strconv.ParseInt(string(text[1:]), 10, 64)
		if err != nil {
			return err
		}
		i.GreaterThan = gt
		return nil
	}
	eq, err := strconv.ParseInt(string(text[1:]), 10, 64)
	if err != nil {
		return err
	}
	i.Equals = eq
	return nil
}

type StringFilter struct {
	Not      bool
	Equals   string
	Contains string
}

func (i *StringFilter) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	if text[0] == '!' {
		i.Not = true
		text = text[1:]
	}
	if text[0] == '~' {
		i.Contains = string(text[1:])
		return nil
	}
	i.Equals = string(text)
	return nil
}
