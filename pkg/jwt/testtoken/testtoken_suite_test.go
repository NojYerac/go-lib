package testtoken_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTToken(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestToken Suite")
}
