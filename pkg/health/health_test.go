package health_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "source.rad.af/libs/go-lib/pkg/health"
	"source.rad.af/libs/go-lib/pkg/log"
)

var passingChecker CheckFn = func(c context.Context) error { return nil }
var failingChecker CheckFn = func(c context.Context) error { return errors.New("mock error") }

var _ = Describe("health", func() {
	var (
		healthChecker Checker
		ctx           context.Context
		cancel        context.CancelFunc
		ready         chan struct{}
	)
	BeforeEach(func() {
		ready = make(chan struct{})
		l := log.NewLogger(&log.Configuration{HumanFrendly: true, LogLevel: "fatal"})
		ctx, cancel = context.WithCancel(l.WithContext(context.Background()))
		healthChecker = NewChecker(&Configuration{CheckInterval: 10 * time.Millisecond}, WithReadyChan(ready))
	})
	JustBeforeEach(func() {
		go func() {
			defer GinkgoRecover()
			Expect(healthChecker.Start(ctx)).To(MatchError(context.Canceled))
		}()
		<-ready
	})
	AfterEach(func() {
		cancel()
	})
	Context("no checks", func() {
		It("returns passed", func() {
			Expect(healthChecker.Passed()).To(BeTrue())
			Expect(healthChecker.String()).To(Equal("[ping] ok"))
		})
	})
	Context("passing", func() {
		BeforeEach(func() {
			healthChecker.Register("passing", passingChecker)
		})
		It("returns passed", func() {
			Expect(healthChecker.Passed()).To(BeTrue())
			Expect(healthChecker.String()).To(And(
				ContainSubstring("[ping] ok"),
				ContainSubstring("[passing] ok"),
			))
		})
	})
	Context("failing", func() {
		BeforeEach(func() {
			healthChecker.Register("failing", failingChecker)
		})
		It("returns failed", func() {
			Expect(healthChecker.Passed()).To(BeFalse())
			Expect(healthChecker.String()).To(And(
				ContainSubstring("[ping] ok"),
				ContainSubstring("[failing] mock error"),
			))
		})
	})
})
