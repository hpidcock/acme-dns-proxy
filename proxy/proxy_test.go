package proxy_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	. "github.com/matthiasng/dns-challenge-proxy/proxy"
)

// mockedProviderHelper is used to control the Present/CleanUp function behaviour
type mockedProviderHelper struct {
	err    error
	domain string
	token  string
	fqdn   string
	value  string
}

func (m *mockedProviderHelper) call(domain, token, fqdn, value string) error {
	m.domain = domain
	m.token = token
	m.fqdn = fqdn
	m.value = value
	return m.err
}

type mockedProvider struct {
	present mockedProviderHelper
	cleanup mockedProviderHelper
}

func (m *mockedProvider) Present(domain, token, fqdn, value string) error {
	return m.present.call(domain, token, fqdn, value)
}

func (m *mockedProvider) CleanUp(domain, token, fqdn, value string) error {
	return m.cleanup.call(domain, token, fqdn, value)
}

var _ = Describe("Proxy", func() {
	Describe("handle", func() {
		var (
			provider    mockedProvider
			accessRules AccessRules
			request     Request
			err         error

			itSouldFailWith = func(s string) {
				It("sould faild", func() {
					Expect(err).To(MatchError(ContainSubstring(s)))
				})
			}
		)

		AfterEach(func() {
			request = Request{}
			accessRules = AccessRules{}
			provider = mockedProvider{}
		})

		JustBeforeEach(func() {
			logger, _ := zap.NewDevelopment()
			proxy := Proxy{
				Logger:      logger,
				AccessRules: accessRules,
				Provider:    &provider,
			}

			err = proxy.Handle(&request)
		})

		// #todo test cleanup
		Describe("present", func() {
			BeforeEach(func() {
				request.Action = "present"
			})

			Context("empty txt fqdn", func() {
				BeforeEach(func() {
					request.Payload = Payload{
						TxtFQDN:  "",
						TxtValue: "123",
					}
				})

				itSouldFailWith("TXT record: FQDN not set")
			})

			Context("empty txt value", func() {
				BeforeEach(func() {
					request.Payload = Payload{
						TxtFQDN:  "test.local",
						TxtValue: "",
					}
				})

				itSouldFailWith("TXT record: value not set")
			})

			Context("unknown fqdn", func() {
				BeforeEach(func() {
					request.Payload = Payload{
						TxtFQDN:  "test.local",
						TxtValue: "123",
					}
				})

				itSouldFailWith(`no access rule for "test.local" found`)
			})

			Context("invalid auth key", func() {
				BeforeEach(func() {
					accessRules = AccessRules{
						&AccessRule{
							Pattern: MustCompilePattern("test.local"),
							Token:   "123",
						},
					}
					request.Token = "abc"
					request.Payload = Payload{
						TxtFQDN:  "test.local",
						TxtValue: "123",
					}
				})

				itSouldFailWith("access denied")
			})

			Context("provider error", func() {
				BeforeEach(func() {
					accessRules = AccessRules{
						&AccessRule{
							Pattern: MustCompilePattern("test.local"),
							Token:   "123",
						},
					}
					request.Token = "123"
					request.Payload = Payload{
						TxtFQDN:  "test.local",
						TxtValue: "123",
					}

					provider.present.err = errors.New("present err test")
				})

				itSouldFailWith("present err test")
			})

			Context("provider error", func() {
				BeforeEach(func() {
					accessRules = AccessRules{
						&AccessRule{
							Pattern: MustCompilePattern("test.local"),
							Token:   "123",
						},
					}
					request.Token = "123"
					request.Payload = Payload{
						TxtFQDN:  "test.local",
						TxtValue: "123",
					}

					provider.present.err = nil
				})

				It("should succeed", func() {
					Expect(err).To(Succeed())
				})

				Specify("provider present call parameter", func() {
					Expect(provider.present.domain).To(Equal(fmt.Sprintf("%s.", request.Payload.TxtFQDN)))
					Expect(provider.present.token).To(Equal(request.Payload.TxtValue))
					Expect(provider.present.fqdn).To(Equal(request.Payload.TxtFQDN))
					Expect(provider.present.value).To(Equal(request.Payload.TxtValue))
				})
			})
		})
	})
})
