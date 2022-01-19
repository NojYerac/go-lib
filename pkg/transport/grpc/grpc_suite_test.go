package grpc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestGRPC(t *testing.T) {
	RegisterFailHandler(Fail)
	reporter := reporters.NewJUnitReporter("report.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "gRPC Suite", []Reporter{reporter})
}
