package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	r := reporters.NewJUnitReporter("report.xml")
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Config Suite", []Reporter{r})
}
