package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/pflag"
)

func NewFlagValue(rv reflect.Value) pflag.Value {
	return &flagValue{
		rv:  rv,
		str: fmt.Sprintf("%v", rv.Interface()),
	}
}

type flagValue struct {
	rv  reflect.Value
	str string
}

func (v *flagValue) String() string {
	return v.str
}

func (v *flagValue) Set(val string) error {
	v.str = val
	return nil
}

func (v *flagValue) Type() string {
	return v.rv.Type().String()
}
