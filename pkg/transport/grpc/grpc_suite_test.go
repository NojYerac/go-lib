package grpc

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestGRPC(t *testing.T) {
	RegisterFailHandler(Fail)
	reporter := reporters.NewJUnitReporter("report.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "gRPC Suite", []Reporter{reporter})
}

var _ = BeforeSuite(func() {
	var b bytes.Buffer
	logrus.SetOutput(&b)
	logrus.SetLevel(logrus.FatalLevel)
})
