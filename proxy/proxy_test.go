package proxy_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/matthiasng/acme-dns-proxy/dns"
	. "github.com/matthiasng/acme-dns-proxy/proxy"
)

type mockProvider struct {
	err       error
	challenge dns.Challenge
}

func (m *mockProvider) Present(c dns.Challenge) error {
	m.challenge = c
	return m.err
}

func (m *mockProvider) CleanUp(c dns.Challenge) error {
	m.challenge = c
	return m.err
}

var _ = Describe("Proxy", func() {
	Describe("handle", func() {
		var (
			provider    mockProvider
			accessRules AccessRules
			request     Request
			err         error
		)

		AfterEach(func() {
			request = Request{}
			accessRules = AccessRules{}
			provider = mockProvider{}
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

		Context("unknown action", func() {
			BeforeEach(func() {
				request.Action = "unknown"
			})

			It("sould fail", func() {
				Expect(err).To(MatchError(ContainSubstring("invalid request: unknown action")))
			})
		})

		for _, action := range []string{"present", "cleanup"} {
			func(action string) {
				Describe(action, func() {
					BeforeEach(func() {
						request.Action = action
					})

					Context("empty txt fqdn", func() {
						BeforeEach(func() {
							request.Challenge = dns.Challenge{
								FQDN:           "",
								EncodedKeyAuth: "123",
							}
						})

						It("sould fail", func() {
							Expect(err).To(MatchError(ContainSubstring("invalid request: fqdn not set")))
						})
					})

					Context("empty txt value", func() {
						BeforeEach(func() {
							request.Challenge = dns.Challenge{
								FQDN:           "test.local",
								EncodedKeyAuth: "",
							}
						})

						It("sould fail", func() {
							Expect(err).To(MatchError(ContainSubstring("invalid request: key auth value not set")))
						})
					})

					Context("unknown fqdn", func() {
						BeforeEach(func() {
							request.Challenge = dns.Challenge{
								FQDN:           "test.local",
								EncodedKeyAuth: "123",
							}
						})

						It("sould fail", func() {
							Expect(err).To(MatchError(ContainSubstring(`no access rule for "test.local" found`)))
						})
					})

					Context("invalid auth key", func() {
						BeforeEach(func() {
							accessRules = AccessRules{
								&AccessRule{
									Pattern: MustCompilePattern("test.local"),
									Token:   "123",
								},
							}
							request.AuthToken = "abc"
							request.Challenge = dns.Challenge{
								FQDN:           "test.local",
								EncodedKeyAuth: "123",
							}
						})

						It("sould fail", func() {
							Expect(err).To(MatchError(ContainSubstring("access denied")))
						})
					})

					Context("provider error", func() {
						BeforeEach(func() {
							accessRules = AccessRules{
								&AccessRule{
									Pattern: MustCompilePattern("test.local"),
									Token:   "123",
								},
							}
							request.AuthToken = "123"
							request.Challenge = dns.Challenge{
								FQDN:           "test.local",
								EncodedKeyAuth: "123",
							}

							provider.err = errors.New("err test")
						})

						It("sould fail", func() {
							Expect(err).To(MatchError(ContainSubstring("err test")))
						})
					})

					Context("provider success", func() {
						BeforeEach(func() {
							accessRules = AccessRules{
								&AccessRule{
									Pattern: MustCompilePattern("test.local"),
									Token:   "123",
								},
							}
							request.AuthToken = "123"
							request.Challenge = dns.Challenge{
								FQDN:           "test.local",
								EncodedKeyAuth: "123",
							}

							provider.err = nil
						})

						It("should succeed", func() {
							Expect(err).To(Succeed())
						})

						Specify("provider call parameter", func() {
							Expect(provider.challenge).To(Equal(request.Challenge))
						})
					})
				})
			}(action)
		}
	})
})
