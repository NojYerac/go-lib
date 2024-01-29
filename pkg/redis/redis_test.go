package redis_test

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mocks "source.rad.af/libs/go-lib/internal/mocks/go-redis"
	"source.rad.af/libs/go-lib/pkg/health"
	. "source.rad.af/libs/go-lib/pkg/redis"
)

var _ = Describe("Client", func() {
	var (
		mockCmdbl *mocks.Cmdable
		h         health.Checker
		c         Client
		err       error
	)
	BeforeEach(func() {
		h = health.NewChecker(&health.Configuration{CheckInterval: time.Second})
		mockCmdbl = mocks.NewCmdable(GinkgoT())
		res := redis.NewStatusCmd("ping")
		mockCmdbl.On("Ping").Return(res)
		c, err = NewClient(&Configuration{}, WithRedisClient(mockCmdbl), WithHealthChecker(h))
		Expect(err).NotTo(HaveOccurred())
	})
	It("is testable", func() {
		Expect(c).NotTo(BeNil())
	})
	It("checks health with a ping", func() {
		ctx, cancel := context.WithCancel(context.Background())
		mockCmdbl.On("Ping").Return(&redis.StatusCmd{})
		go func() {
			defer GinkgoRecover()
			Expect(h.Start(ctx)).To(Equal(context.Canceled))
		}()
		time.Sleep(20 * time.Millisecond)
		cancel()
		Expect(h.Passed()).To(BeTrue())
	})
	It("Del", func() {
		mockCmdbl.On("Del", "a").Return(&redis.IntCmd{}, nil)
		val, err := c.Del(context.Background(), "a")
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(Equal(int64(0)))
	})
	It("Get", func() {
		mockCmdbl.On("Get", "a").Return(&redis.StringCmd{}, nil)
		val, err := c.Get(context.Background(), "a")
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(Equal(""))
	})
	It("Set", func() {
		mockCmdbl.On("Set", "a", "b", time.Second).Return(&redis.StatusCmd{}, nil)
		err := c.Set(context.Background(), "a", "b", time.Second)
		Expect(err).NotTo(HaveOccurred())
	})
})
