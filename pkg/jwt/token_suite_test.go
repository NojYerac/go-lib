package token

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestExample(t *testing.T) {
	RegisterFailHandler(Fail)
	reporter := reporters.NewJUnitReporter("report.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Example Suite", []Reporter{reporter})
}
