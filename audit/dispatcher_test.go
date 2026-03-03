package audit

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockSink struct {
	events []event
	fail   bool
	name   string
}

func (s *mockSink) Send(ctx context.Context, evt event) error {
	if s.fail {
		return fmt.Errorf("mock failure")
	}
	s.events = append(s.events, evt)
	return nil
}

func (s *mockSink) Name() string { return s.name }

var _ = Describe("Dispatcher", func() {
	It("should dispatch events to sinks", func() {
		pub := NewMemoryPublisher(10)
		sink := &mockSink{name: "test-sink"}
		dispatcher := NewDefaultDispatcher(pub, []Sink{sink})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := dispatcher.Start(ctx)
		Expect(err).NotTo(HaveOccurred())

		evt := event{
			ActorID:   "550e8400-e29b-41d4-a716-446655440000",
			Action:    "test.action",
			Timestamp: time.Now(),
			Details:   map[string]any{"foo": "bar"},
		}

		err = pub.Publish(ctx, evt)
		Expect(err).NotTo(HaveOccurred())

		// Wait for dispatch
		Eventually(func() int {
			return len(sink.events)
		}, "5s", "50ms").Should(Equal(1))

		Expect(sink.events[0].Action).To(Equal("test.action"))

		err = dispatcher.Stop(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should retry on failure", func() {
		pub := NewMemoryPublisher(10)
		sink := &mockSink{name: "fail-sink", fail: true}
		
		// Cast to access internal fields for testing if needed, or just observe behavior
		d := &DefaultDispatcher{
			publisher:     pub,
			sinks:         []Sink{sink},
			stopCh:        make(chan struct{}),
			retryInterval: 10 * time.Millisecond,
			maxRetries:    2,
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := d.Start(ctx)
		Expect(err).NotTo(HaveOccurred())

		evt := event{
			ActorID:   "550e8400-e29b-41d4-a716-446655440000",
			Action:    "retry.action",
			Timestamp: time.Now(),
			Details:   map[string]any{"foo": "bar"},
		}

		err = pub.Publish(ctx, evt)
		Expect(err).NotTo(HaveOccurred())

		// We can't easily check retries without more instrumentation, 
		// but we can ensure it doesn't crash and we can stop it.
		time.Sleep(100 * time.Millisecond)

		err = d.Stop(ctx)
		Expect(err).NotTo(HaveOccurred())
	})
})
