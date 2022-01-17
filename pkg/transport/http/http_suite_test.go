package http_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestHTTP(t *testing.T) {
	r := reporters.NewJUnitReporter("report.xml")
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "HTTP Suite", []Reporter{r})
}
