package version_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestVersion(t *testing.T) {
	r := reporters.NewJUnitReporter("report.xml")
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Version Suite", []Reporter{r})
}
