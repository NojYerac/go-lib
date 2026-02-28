package audit

import "time"

type Actor struct {
	Type string
	ID   string
	IP   string
}

type Resource struct {
	Type string
	ID   string
}

type Event struct {
	ID        string
	Action    string
	Actor     Actor
	Resource  Resource
	Timestamp time.Time
	Details   map[string]any
}
