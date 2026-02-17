package config_test

import (
	"testing"

	. "github.com/nojyerac/go-lib/pkg/config"
	"github.com/nojyerac/go-lib/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var c Loader
var _ = BeforeSuite(func() {
	l := log.NewLogger(&log.Configuration{HumanFrendly: true, LogLevel: "fatal"})
	c = NewConfigLoader("test", WithLogger(l), WithArgs("-c", "./testdata"))
})
