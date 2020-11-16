package dns_test

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/libdns/libdns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/matthiasng/acme-dns-proxy/config"
	. "github.com/matthiasng/acme-dns-proxy/dns"
)

type mockLibdnsProvider struct {
	records   []libdns.Record
	zone      string
	appendErr error
	deleteErr error
}

func (p *mockLibdnsProvider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.records = recs
	p.zone = zone

	result := recs
	for i := 0; i < len(recs); i++ {
		result[i].ID = uuid.New().String()
	}

	return p.records, p.appendErr
}

func (p *mockLibdnsProvider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.records = recs
	p.zone = zone
	return p.records, p.deleteErr
}

func (p *mockLibdnsProvider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	// unused
	return nil, nil
}

func (p *mockLibdnsProvider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	// unused
	return nil, nil
}

func mockZoneResolver(zone string, err error) ZoneResolver {
	return func(_ string) (string, error) {
		return zone, err
	}
}

var _ = Describe("Provider", func() {
	Describe("from config", func() {
		var (
			providerCfg config.Provider
			provider    Provider
			err         error
		)

		JustBeforeEach(func() {
			provider, err = NewProviderFromConfig(&providerCfg, nil)
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
			It("sould fail", func() {
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
			It("sould fail", func() {
				Expect(err).To(MatchError(ContainSubstring("error initializing provider: Unknown provider: unknown-type")))
			})
		})

		Context("with missing provider variables", func() {
			BeforeEach(func() {
				providerCfg = config.Provider{
					Type:      "hetzner",
					Variables: map[string]string{"UNKNOWN": "ABC"},
				}
			})

			It("should return nil", func() {
				Expect(provider).To(BeNil())
			})
			It("sould fail", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring(`"AuthAPIToken" not set`)))
			})
		})

		Context("with valid provider config", func() {
			BeforeEach(func() {
				providerCfg = config.Provider{
					Type:      "hetzner",
					Variables: map[string]string{"AuthAPIToken": "ABC"},
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

	Describe("present", func() {
		var (
			mockLibdns   *mockLibdnsProvider
			zoneResolver ZoneResolver
			challenge    Challenge

			provider Provider
			err      error
		)

		BeforeEach(func() {
			mockLibdns = &mockLibdnsProvider{}
			zoneResolver = mockZoneResolver("", nil)
		})

		JustBeforeEach(func() {
			provider, _ = NewProvider(mockLibdns, zoneResolver)
			err = provider.Present(challenge)
		})

		Context("zone resolver error", func() {
			BeforeEach(func() {
				zoneResolver = mockZoneResolver("", errors.New("zone resolver err"))
			})

			It("sould fail", func() {
				Expect(err).To(MatchError(ContainSubstring("zone resolver err")))
			})
		})

		Context("provider error", func() {
			BeforeEach(func() {
				mockLibdns = &mockLibdnsProvider{
					appendErr: errors.New("provider err"),
				}
			})

			It("sould fail", func() {
				Expect(err).To(MatchError(ContainSubstring("provider err")))
			})
		})

		Context("valid challenge", func() {
			BeforeEach(func() {
				challenge = Challenge{FQDN: "example.com", EncodedKeyAuth: "123"}
			})

			It("should succeed", func() {
				Expect(err).To(Succeed())
			})
		})
	})

	Describe("cleanup", func() {
		var (
			mockLibdns       *mockLibdnsProvider
			zoneResolver     ZoneResolver
			challenge        Challenge
			cleanupChallenge Challenge

			provider Provider
			err      error
		)

		BeforeEach(func() {
			mockLibdns = &mockLibdnsProvider{}
			zoneResolver = mockZoneResolver("", nil)
			challenge = Challenge{FQDN: "example.com", EncodedKeyAuth: "123"}
			cleanupChallenge = Challenge{FQDN: "example.com", EncodedKeyAuth: "123"}
		})

		JustBeforeEach(func() {
			provider, _ = NewProvider(mockLibdns, zoneResolver)
			_ = provider.Present(challenge)
			err = provider.CleanUp(cleanupChallenge)
		})

		Context("without present", func() {
			BeforeEach(func() {
				cleanupChallenge = Challenge{FQDN: "example.com", EncodedKeyAuth: "unknown"}
			})

			It("sould fail", func() {
				Expect(err).To(MatchError(ContainSubstring("no pending record found")))
			})
		})

		Context("provider error", func() {
			BeforeEach(func() {
				mockLibdns = &mockLibdnsProvider{
					deleteErr: errors.New("provider err"),
				}
			})

			It("sould fail", func() {
				Expect(err).To(MatchError(ContainSubstring("provider err")))
			})
		})

		Context("valid challenge", func() {
			BeforeEach(func() {
				challenge = Challenge{FQDN: "example.com", EncodedKeyAuth: "123"}
			})

			It("should succeed", func() {
				Expect(err).To(Succeed())
			})
		})
	})
})
