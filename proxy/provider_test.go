package proxy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/matthiasng/dns-challenge-proxy/config"
	. "github.com/matthiasng/dns-challenge-proxy/proxy"
)

var _ = Describe("Provider", func() {
	Describe("from config", func() {

		var (
			providerCfg config.Provider
			provider    Provider
			err         error
		)

		JustBeforeEach(func() {
			provider, err = NewLegoProviderFromConfig(&providerCfg)
		})

		Context("with empty type", func() {
			BeforeEach(func() {
				providerCfg = config.Provider{
					Type: "",
				}
			})

			It("should return nil", func() {
				Expect(provider).To(BeNil())
			})
			It("sould faild", func() {
				Expect(err).To(MatchError(ContainSubstring("provider type not specified")))
			})
		})

		Context("with unknown type", func() {
			BeforeEach(func() {
				providerCfg = config.Provider{
					Type: "unknown-type",
				}
			})

			It("should return nil", func() {
				Expect(provider).To(BeNil())
			})
			It("sould faild", func() {
				Expect(err).To(MatchError(ContainSubstring("unrecognized DNS provider: unknown-type")))
			})
		})

		Context("with missing provider variables", func() {
			BeforeEach(func() {
				providerCfg = config.Provider{
					Type:      "digitalocean",
					Variables: map[string]string{"UNKNOWN": "ABC"},
				}
			})

			It("should return nil", func() {
				Expect(provider).To(BeNil())
			})
			It("sould faild", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("digitalocean: some credentials information are missing: DO_AUTH_TOKEN")))
			})
		})

		Context("with valid provider config", func() {
			BeforeEach(func() {
				providerCfg = config.Provider{
					Type:      "digitalocean",
					Variables: map[string]string{"DO_AUTH_TOKEN": "ABC"},
				}
			})

			It("should succeed", func() {
				Expect(err).To(Succeed())
			})
			It("sould return a provider", func() {
				Expect(provider).To(Not(BeNil()))
			})
		})
	})

})
