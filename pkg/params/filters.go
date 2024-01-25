package params

import (
	"encoding"
	"strconv"
)

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
