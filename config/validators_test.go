package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TestConfiguration struct {
	PrivateKey string `validate:"priv_ec_key"`
	PublicKey  string `validate:"pub_key"`
}

var testConfig = &TestConfiguration{
	PrivateKey: `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIA4WF79lBYQCjjIOunx5N75WdqstUwI4XYIqLZSxyJtqoAoGCCqGSM49
AwEHoUQDQgAEyViEkF0WSOVwYcISC9bokDxVVibYftwFC/YY3Q3oXDX0iAD3waIm
4J9yN4gD1K8kmQN80jfjSjf2k2hLhZ1X3Q==
-----END EC PRIVATE KEY-----`,
	PublicKey: `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEyViEkF0WSOVwYcISC9bokDxVVibY
ftwFC/YY3Q3oXDX0iAD3waIm4J9yN4gD1K8kmQN80jfjSjf2k2hLhZ1X3Q==
-----END PUBLIC KEY-----`,
}

var _ = Describe("validators", func() {
	It("validates", func() {
		Expect(c.RegisterConfig(testConfig)).To(Succeed())
		Expect(c.InitAndValidate()).To(Succeed())
	})
})
