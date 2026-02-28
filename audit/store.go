package audit

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Writer interface {
	Append(context.Context, *Event, AppendOptions) error
}

type Reader interface {
	List(context.Context, *ListOptions) (ListResult, error)
}

type AppendOptions struct {
	TransactionID string
}

type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

type Query struct {
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Since        *time.Time
	Until        *time.Time
}

type Page struct {
	Limit  int
	Cursor string
	Order  Order
}

type ListOptions struct {
	Filter Query
	Page   Page
}

type PageInfo struct {
	NextCursor string
	HasMore    bool
}

type ListResult struct {
	Events   []Event
	PageInfo PageInfo
}

type record struct {
	seq           uint64
	event         Event
	transactionID string
}

type MemoryStore struct {
	mu      sync.RWMutex
	events  []record
	nextSeq uint64
	cfg     *Configuration
}

func NewMemoryStore(cfg *Configuration) *MemoryStore {
	if cfg == nil {
		cfg = NewConfiguration()
	}

	c := *cfg
	return &MemoryStore{cfg: &c, nextSeq: 1}
}

func (m *MemoryStore) Append(_ context.Context, event *Event, opts AppendOptions) error {
	if err := ValidateEvent(event, m.cfg); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, record{
		seq: m.nextSeq,
		event: Event{
			ID:        event.ID,
			Action:    event.Action,
			Actor:     event.Actor,
			Resource:  event.Resource,
			Timestamp: event.Timestamp,
			Details:   cloneMap(event.Details),
		},
		transactionID: opts.TransactionID,
	})
	m.nextSeq++

	return nil
}

func (m *MemoryStore) List(_ context.Context, opts *ListOptions) (ListResult, error) {
	if opts == nil {
		opts = &ListOptions{}
	}

	order := normalizeOrder(opts.Page.Order)
	limit := m.cfg.normalizePageLimit(opts.Page.Limit)
	cursor, err := parseCursor(opts.Page.Cursor)
	if err != nil {
		return ListResult{}, err
	}

	filtered := m.filteredRecords(&opts.Filter)
	selected := selectRecords(filtered, order, cursor, limit)
	return buildResult(selected, limit), nil
}

func (m *MemoryStore) filteredRecords(query *Query) []record {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filtered := make([]record, 0, len(m.events))
	for idx := range m.events {
		current := m.events[idx]
		if !matches(&current.event, query) {
			continue
		}
		filtered = append(filtered, current)
	}

	return filtered
}

func selectRecords(records []record, order Order, cursor uint64, limit int) []record {
	selected := make([]record, 0, limit+1)
	if order == OrderDesc {
		for idx := len(records) - 1; idx >= 0; idx-- {
			current := records[idx]
			if cursor > 0 && current.seq >= cursor {
				continue
			}
			selected = append(selected, current)
			if len(selected) > limit {
				return selected
			}
		}
		return selected
	}

	for idx := range records {
		current := records[idx]
		if cursor > 0 && current.seq <= cursor {
			continue
		}
		selected = append(selected, current)
		if len(selected) > limit {
			return selected
		}
	}

	return selected
}

func buildResult(selected []record, limit int) ListResult {
	hasMore := len(selected) > limit
	if hasMore {
		selected = selected[:limit]
	}

	events := make([]Event, 0, len(selected))
	for idx := range selected {
		current := selected[idx]
		events = append(events, Event{
			ID:        current.event.ID,
			Action:    current.event.Action,
			Actor:     current.event.Actor,
			Resource:  current.event.Resource,
			Timestamp: current.event.Timestamp,
			Details:   cloneMap(current.event.Details),
		})
	}

	result := ListResult{
		Events: events,
		PageInfo: PageInfo{
			HasMore: hasMore,
		},
	}
	if hasMore && len(selected) > 0 {
		result.PageInfo.NextCursor = formatCursor(selected[len(selected)-1].seq)
	}

	return result
}

func parseCursor(cursor string) (uint64, error) {
	if cursor == "" {
		return 0, nil
	}

	value, err := strconv.ParseUint(cursor, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidCursor, cursor)
	}

	return value, nil
}

func formatCursor(value uint64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatUint(value, 10)
}

func normalizeOrder(order Order) Order {
	if order == OrderDesc {
		return OrderDesc
	}
	return OrderAsc
}

func matches(event *Event, query *Query) bool {
	if event == nil {
		return false
	}
	if query == nil {
		return true
	}

	if query.ActorID != "" && event.Actor.ID != query.ActorID {
		return false
	}

	if query.Action != "" && event.Action != query.Action {
		return false
	}

	if query.ResourceType != "" && event.Resource.Type != query.ResourceType {
		return false
	}

	if query.ResourceID != "" && event.Resource.ID != query.ResourceID {
		return false
	}

	if query.Since != nil && event.Timestamp.Before(*query.Since) {
		return false
	}

	if query.Until != nil && event.Timestamp.After(*query.Until) {
		return false
	}

	return true
}

func cloneMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}

	return cloned
}
