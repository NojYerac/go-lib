package audit_test

import (
	"errors"
	"time"

	. "github.com/nojyerac/go-lib/audit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validation", func() {
	newEvent := func() *Event {
		return &Event{
			ID:     "evt-1",
			Action: "resource.updated",
			Actor: Actor{
				Type: "user",
				ID:   "u-1",
			},
			Resource: Resource{
				Type: "order",
				ID:   "o-1",
			},
			Timestamp: time.Now().UTC(),
			Details: map[string]any{
				"changed": true,
			},
		}
	}

	It("accepts a valid event", func() {
		Expect(ValidateEvent(newEvent(), nil)).To(Succeed())
	})

	It("rejects nil event", func() {
		err := ValidateEvent(nil, nil)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidEvent)).To(BeTrue())
	})

	It("rejects missing action", func() {
		event := newEvent()
		event.Action = ""
		err := ValidateEvent(event, nil)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidAction)).To(BeTrue())
	})

	It("rejects missing actor identity", func() {
		event := newEvent()
		event.Actor.ID = ""
		err := ValidateEvent(event, nil)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidActor)).To(BeTrue())
	})

	It("rejects missing resource identity", func() {
		event := newEvent()
		event.Resource.Type = ""
		err := ValidateEvent(event, nil)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidResource)).To(BeTrue())
	})

	It("rejects zero timestamp", func() {
		event := newEvent()
		event.Timestamp = time.Time{}
		err := ValidateEvent(event, nil)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidTimestamp)).To(BeTrue())
	})

	It("rejects details over configured payload limit", func() {
		event := newEvent()
		event.Details = map[string]any{"blob": "this payload is too large"}
		cfg := NewConfiguration()
		cfg.MaxDetailsBytes = 8

		err := ValidateEvent(event, cfg)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrDetailsTooLarge)).To(BeTrue())
	})
})

var _ = Describe("MarshalBoundedJSON", func() {
	It("returns invalid limit error for negative limit", func() {
		_, err := MarshalBoundedJSON(map[string]any{"k": "v"}, -1)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidLimit)).To(BeTrue())
	})

	It("returns details too large when marshaled payload exceeds max", func() {
		_, err := MarshalBoundedJSON(map[string]any{"blob": "1234567890"}, 5)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrDetailsTooLarge)).To(BeTrue())
	})

	It("allows unlimited payload when maxBytes is zero", func() {
		payload, err := MarshalBoundedJSON(map[string]any{"blob": "1234567890"}, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(payload).NotTo(BeEmpty())
	})
})
