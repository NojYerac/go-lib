package params

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	apierrors "github.com/go-openapi/errors"
	"github.com/go-playground/validator/v10"
)

// parameters
const (
	query        = "query"
	orderByParam = "orderBy"
	limitParam   = "limit"
	offsetParam  = "offset"
)

var (
	cache               = map[reflect.Type]map[*paramTag]setterFunc{}
	trueValues          = []string{"true", "TRUE", "t", "1"}
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	v                   = validator.New()
)

type setterFunc func(string) (reflect.Value, error)

type validater interface {
	Validate() error
}

type paramTag struct {
	Name      string
	Modifiers []string
	Index     int
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
	for paramTag, setter := range makeSetters(rte) {
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

func makeSetters(rt reflect.Type) map[*paramTag]setterFunc {
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
		setters[tagVal] = makeSetter(field.Type, tagVal.Modifiers)
	}
	cache[rt] = setters
	return setters
}

func makeSetter(rt reflect.Type, mods []string) setterFunc {
	var _ = len(mods) // may need mods to tweak setters

	if rt.Implements(textUnmarshalerType) {
		return makeUnmarshalerSetter(rt)
	}
	switch rt.Kind() {
	case reflect.Ptr:
		pSetter := makeSetter(rt.Elem(), mods)
		return makePtrSetter(rt, pSetter)
	case reflect.Slice:
		rte := rt.Elem()
		iSetter := makeSetter(rte, mods)
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
		return func(val string) (rv reflect.Value, err error) {
			err = fmt.Errorf("unsupported type %s", rt.String())
			return
		}
	}
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
