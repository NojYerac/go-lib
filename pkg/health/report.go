package health

import (
	"fmt"
	"strings"
	"sync"
)

func newReport(err error) Report {
	return &report{err}
}

type Report interface {
	Passed() bool
	String() string
}

type report struct {
	err error
}

func (r *report) Passed() bool {
	return r.err == nil
}

func (r *report) String() string {
	if r.err != nil {
		return r.err.Error()
	}
	return "ok"
}

type reports struct {
	sync.RWMutex
	m map[string]Report
}

func (r *reports) Passed() bool {
	r.RLock()
	defer r.RUnlock()
	for _, sub := range r.m {
		if !sub.Passed() {
			return false
		}
	}
	return true
}

func (r *reports) String() string {
	r.RLock()
	defer r.RUnlock()
	strs := make([]string, 0, len(r.m))
	for name, sub := range r.m {
		strs = append(strs, fmt.Sprint("[", name, "] ", sub.String()))
	}
	return strings.Join(strs, "\n")
}
