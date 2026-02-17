package testtoken_test

import (
	"time"

	gojwt "github.com/dgrijalva/jwt-go"
	"github.com/nojyerac/go-lib/pkg/jwt"
	. "github.com/nojyerac/go-lib/pkg/jwt/testtoken"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func expired() time.Time {
	return time.Now().Add(24 * time.Hour)
}

func notYet() time.Time {
	return time.Now().Add(-24 * time.Hour)
}

var _ = Describe("Testtoken", func() {

	It("is testable", func() {
		Expect(true).To(BeTrue())
	})
	It("validates as expected", func() {
		var (
			session *jwt.Session
			err     error
		)
		testToken := TestToken()

		gojwt.TimeFunc = expired
		session, err = Issuer.Session(testToken)
		Expect(err).To(MatchError("token is expired by 23h59m0s"))
		Expect(session).To(BeNil())

		gojwt.TimeFunc = notYet
		session, err = Issuer.Session(testToken)
		Expect(err).To(MatchError("token is not valid yet"))
		Expect(session).To(BeNil())

		gojwt.TimeFunc = time.Now
		session, err = Issuer.Session(testToken)
		Expect(err).NotTo(HaveOccurred())
		Expect(session).NotTo(BeNil())
	})
})
