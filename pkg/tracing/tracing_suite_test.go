package tracing_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestTracing(t *testing.T) {
	r := reporters.NewJUnitReporter("report.xml")
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Tracing Suite", []Reporter{r})
}
