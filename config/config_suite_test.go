package config_test

import (
	"testing"

	. "github.com/nojyerac/go-lib/config"

	"github.com/nojyerac/go-lib/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var c Loader
var _ = BeforeSuite(func() {
	c = NewConfigLoader("test", WithLogger(log.Nop()), WithArgs("-c", "./testdata"))
})
