package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "source.rad.af/libs/go-lib/pkg/config"
	"source.rad.af/libs/go-lib/pkg/log"
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
