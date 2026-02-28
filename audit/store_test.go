package audit_test

import (
	"context"
	"errors"
	"strconv"
	"time"

	. "github.com/nojyerac/go-lib/audit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemoryStore", func() {
	newEvent := func(id string, when time.Time, action string, actorID string, resourceID string) *Event {
		return &Event{
			ID:     id,
			Action: action,
			Actor: Actor{
				Type: "user",
				ID:   actorID,
			},
			Resource: Resource{
				Type: "order",
				ID:   resourceID,
			},
			Timestamp: when,
			Details: map[string]any{
				"source": "test",
			},
		}
	}

	It("appends and lists in ascending order by default", func() {
		store := NewMemoryStore(nil)
		now := time.Now().UTC()
		Expect(
			store.Append(
				context.Background(),
				newEvent("1", now, "a", "u1", "r1"),
				AppendOptions{},
			),
		).To(Succeed())
		Expect(
			store.Append(
				context.Background(),
				newEvent("2", now.Add(time.Second), "b", "u1", "r2"),
				AppendOptions{},
			),
		).To(Succeed())

		result, err := store.List(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Events).To(HaveLen(2))
		Expect(result.Events[0].ID).To(Equal("1"))
		Expect(result.Events[1].ID).To(Equal("2"))
		Expect(result.PageInfo.HasMore).To(BeFalse())
	})

	It("supports descending pagination with cursor", func() {
		store := NewMemoryStore(nil)
		now := time.Now().UTC()
		for i := 1; i <= 4; i++ {
			id := strconv.Itoa(i)
			Expect(
				store.Append(
					context.Background(),
					newEvent(id, now.Add(time.Duration(i)*time.Second), "a", "u1", id),
					AppendOptions{},
				),
			).To(Succeed())
		}

		firstPage, err := store.List(context.Background(), &ListOptions{Page: Page{Limit: 2, Order: OrderDesc}})
		Expect(err).NotTo(HaveOccurred())
		Expect(firstPage.Events).To(HaveLen(2))
		Expect(firstPage.Events[0].ID).To(Equal("4"))
		Expect(firstPage.Events[1].ID).To(Equal("3"))
		Expect(firstPage.PageInfo.HasMore).To(BeTrue())
		Expect(firstPage.PageInfo.NextCursor).NotTo(BeEmpty())

		secondPage, err := store.List(context.Background(), &ListOptions{
			Page: Page{Limit: 2, Order: OrderDesc, Cursor: firstPage.PageInfo.NextCursor},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(secondPage.Events).To(HaveLen(2))
		Expect(secondPage.Events[0].ID).To(Equal("2"))
		Expect(secondPage.Events[1].ID).To(Equal("1"))
		Expect(secondPage.PageInfo.HasMore).To(BeFalse())
	})

	It("applies filter and time bounds", func() {
		store := NewMemoryStore(nil)
		now := time.Now().UTC()

		Expect(
			store.Append(
				context.Background(),
				newEvent("1", now, "order.create", "u1", "r1"),
				AppendOptions{},
			),
		).To(Succeed())
		Expect(
			store.Append(
				context.Background(),
				newEvent("2", now.Add(time.Minute), "order.update", "u2", "r1"),
				AppendOptions{},
			),
		).To(Succeed())
		Expect(
			store.Append(
				context.Background(),
				newEvent("3", now.Add(2*time.Minute), "order.update", "u1", "r2"),
				AppendOptions{},
			),
		).To(Succeed())

		since := now.Add(30 * time.Second)
		until := now.Add(90 * time.Second)
		result, err := store.List(context.Background(), &ListOptions{Filter: Query{
			Action:     "order.update",
			ActorID:    "u2",
			ResourceID: "r1",
			Since:      &since,
			Until:      &until,
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Events).To(HaveLen(1))
		Expect(result.Events[0].ID).To(Equal("2"))
	})

	It("returns invalid cursor error", func() {
		store := NewMemoryStore(nil)
		_, err := store.List(context.Background(), &ListOptions{Page: Page{Cursor: "nope"}})
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidCursor)).To(BeTrue())
	})

	It("returns invalid event error when append event is nil", func() {
		store := NewMemoryStore(nil)
		err := store.Append(context.Background(), nil, AppendOptions{})
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidEvent)).To(BeTrue())
	})

	It("clones details on write and read", func() {
		store := NewMemoryStore(nil)
		now := time.Now().UTC()
		event := newEvent("1", now, "order.update", "u1", "r1")
		Expect(store.Append(context.Background(), event, AppendOptions{})).To(Succeed())

		event.Details["source"] = "mutated"
		result, err := store.List(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Events).To(HaveLen(1))
		Expect(result.Events[0].Details["source"]).To(Equal("test"))

		result.Events[0].Details["source"] = "changed"
		resultAgain, err := store.List(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(resultAgain.Events[0].Details["source"]).To(Equal("test"))
	})

	It("enforces max page size from configuration", func() {
		cfg := NewConfiguration()
		cfg.DefaultPageSize = 1
		cfg.MaxPageSize = 2
		store := NewMemoryStore(cfg)
		now := time.Now().UTC()

		for i := 1; i <= 3; i++ {
			id := strconv.Itoa(i)
			Expect(
				store.Append(
					context.Background(),
					newEvent(id, now.Add(time.Duration(i)*time.Second), "a", "u1", id),
					AppendOptions{},
				),
			).To(Succeed())
		}

		result, err := store.List(context.Background(), &ListOptions{Page: Page{Limit: 100}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Events).To(HaveLen(2))
		Expect(result.PageInfo.HasMore).To(BeTrue())
	})
})
