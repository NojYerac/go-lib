package params_test

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/nojyerac/go-lib/pkg/params"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testFilter struct {
	Strings   []string      `params:"strings,modify"`
	Ints      []int         `params:"ints"`
	StrPtr    *string       `params:"strPtr"`
	Dur       time.Duration `params:"duration"`
	Time      time.Time     `params:"time"`
	Omit      int           `params:"-"`
	Unmarsher []*IntFilter  `params:"tu"`
	NoTag1    uint          `validate:"lt=10"`
	NoTag2    bool
}

var _ = Describe("OrderBy", func() {
	var (
		o OrderBy
		s string
	)
	JustBeforeEach(func() {
		s = o.String()
	})
	BeforeEach(func() {
		o.Column = "column"
	})
	Context("with column, but no dir", func() {
		It("returns ascending", func() {
			Expect(s).To(Equal("column ASC"))
		})
	})
	Context("with column and dir", func() {
		BeforeEach(func() {
			o.Dir = true
		})
		It("returns descending", func() {
			Expect(s).To(Equal("column DESC"))
		})
	})
})

var _ = Describe("GetPage", func() {
	var (
		c    *gin.Context
		page *Page
		uri  string
		err  error
	)
	JustBeforeEach(func() {
		req, _ := http.NewRequest("GET", uri, http.NoBody)
		c = &gin.Context{Request: req}
		page, err = GetPage(c)
	})
	It("should not return an error", func() {
		Expect(err).NotTo(HaveOccurred())
	})
	Context("with valid values", func() {
		BeforeEach(func() {
			uri = "/?limit=99&offset=999"
		})
		It("should populate the struct", func() {
			Expect(page).NotTo(BeNil())
			Expect(page).To(Equal(&Page{
				Limit:  99,
				Offset: 999,
			}))
		})
	})
	Context("with invalid values", func() {
		BeforeEach(func() {
			uri = "/?limit=a&offset=999"
		})
		It("return an error", func() {
			Expect(page).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("GetOrderBy", func() {
	var (
		c   *gin.Context
		ob  []*OrderBy
		uri string
		err error
	)
	JustBeforeEach(func() {
		req, _ := http.NewRequest("GET", uri, http.NoBody)
		c = &gin.Context{Request: req}
		ob, err = GetOrderBy(c)
	})
	It("should not return an error", func() {
		Expect(err).NotTo(HaveOccurred())
	})
	Context("with valid values", func() {
		BeforeEach(func() {
			uri = "/?orderBy=col1,-col2"
		})
		It("should populate the struct", func() {
			Expect(ob).To(HaveLen(2))
			Expect(ob[0]).To(Equal(&OrderBy{
				Column: "col1",
				Dir:    ASC,
			}))
			Expect(ob[1]).To(Equal(&OrderBy{
				Column: "col2",
				Dir:    DESC,
			}))
		})
	})
	Context("with invalid values", func() {
		BeforeEach(func() {
			uri = "/?orderBy=*col"
		})
		It("return an error", func() {
			Expect(ob).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("GetFilters", func() {
	var (
		c   *gin.Context
		tf  *testFilter
		uri string
		err error
	)
	JustBeforeEach(func() {
		tf = &testFilter{}
		req, _ := http.NewRequest("GET", uri, http.NoBody)
		c = &gin.Context{Request: req}
		err = GetFilters(tf, c)
	})
	It("should not return an error", func() {
		Expect(err).NotTo(HaveOccurred())
	})
	Context("with valid values", func() {
		BeforeEach(func() {
			uri = "/?strings=this,that&ints=1,99&strPtr=abc&duration=30s" +
				"&time=2020-01-01T00:00:00Z&notag1=3&notag2=true&tu=>0,!<10"
		})
		It("should populate the struct", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(tf.Strings).To(ConsistOf("this", "that"))
			Expect(tf.Ints).To(ConsistOf(1, 99))
			expectedStrPtr := "abc"
			Expect(tf.StrPtr).To(Equal(&expectedStrPtr))
			Expect(tf.Dur).To(BeNumerically("==", 30*time.Second))
			expectedTime, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
			Expect(tf.Time).To(BeTemporally("==", expectedTime))
			Expect(tf.NoTag1).To(Equal(uint(3)))
			Expect(tf.NoTag2).To(BeTrue())
			Expect(tf.Unmarsher).To(And(
				HaveLen(2),
				ConsistOf(&IntFilter{GreaterThan: 0}, &IntFilter{Not: true, LessThan: 10}),
			))
		})
	})
	Context("with wrongly typed values", func() {
		BeforeEach(func() {
			uri = "/?ints=a"
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("with invalid values", func() {
		BeforeEach(func() {
			uri = "/?notag1=99"
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("caching", func() {
		It("caches the setters", func() {
			// Expect(cache[reflect.TypeOf(*tf)]).To(HaveLen(7))
		})
	})
})
