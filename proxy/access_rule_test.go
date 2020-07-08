package proxy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/matthiasng/acme-dns-proxy/config"
	. "github.com/matthiasng/acme-dns-proxy/proxy"
)

var _ = Describe("AccessRules", func() {
	Describe("from config", func() {
		var (
			rulesCfg config.AccessRules
			rules    AccessRules
			err      error
		)

		JustBeforeEach(func() {
			rules, err = NewAccessRulesFromConfig(rulesCfg)
		})

		Context("with empty config", func() {
			BeforeEach(func() {
				rulesCfg = config.AccessRules{}
			})

			It("should return nil", func() {
				Expect(rules).To(BeNil())
			})
			It("sould faild", func() {
				Expect(err).To(MatchError(ContainSubstring("no access rules defined")))
			})
		})

		Context("with rule without token", func() {
			BeforeEach(func() {
				rulesCfg = config.AccessRules{
					{Pattern: "abc", Token: ""},
				}
			})

			It("should return nil", func() {
				Expect(rules).To(BeNil())
			})
			It("sould faild", func() {
				Expect(err).To(MatchError(ContainSubstring("'token' not specified")))
			})
		})

		Context("with 3 rules", func() {
			BeforeEach(func() {
				rulesCfg = config.AccessRules{
					{Pattern: "abc.test", Token: "123"},
					{Pattern: "foo.dev", Token: "bar"},
					{Pattern: "*.test.local", Token: "secret"},
				}
			})

			It("should succeed", func() {
				Expect(err).To(Succeed())
			})
			Specify("3 rules", func() {
				Expect(rules).To(HaveLen(3))
			})
			It("should find 'abc.test.local'", func() {
				rule := rules.Search("abc.test.local")
				Expect(rule).To(Not(BeNil()))
				Expect(rule.Pattern.String()).To(Equal("*.test.local"))
				Expect(rule.Token).To(Equal("secret"))
			})
			It("should not find 'test.unknown'", func() {
				rule := rules.Search("test.unknown")
				Expect(rule).To(BeNil())
			})
		})
	})
})

var _ = Describe("AccessRule", func() {
	Context("check auth key", func() {
		pattern := MustCompilePattern("")
		rule := AccessRule{
			Pattern: pattern,
			Token:   "1234567890",
		}

		It("sould be false with invalid key", func() {
			Expect(rule.CheckAuth("abc")).To(BeFalse())
		})
		It("sould be true with valid key", func() {
			Expect(rule.CheckAuth("1234567890")).To(BeTrue())
		})
	})
})
