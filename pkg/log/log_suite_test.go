package log_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestLog(t *testing.T) {
	r := reporters.NewJUnitReporter("report.xml")
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Log Suite", []Reporter{r})
}
