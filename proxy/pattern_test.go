package proxy_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/matthiasng/acme-dns-proxy/proxy"
)

var _ = Describe("Pattern", func() {
	var (
		source  string
		pattern *Pattern
		err     error
	)

	JustBeforeEach(func() {
		pattern, err = CompilePattern(source)
	})

	Context("'*.test'", func() {
		BeforeEach(func() {
			source = "*.test"
		})

		It("should compile succesfully", func() {
			Expect(err).To(Succeed())
		})
		It("sould return a pattern", func() {
			Expect(pattern).To(Not(BeNil()))
			Expect(pattern.String()).To(Equal("*.test"))
		})

		for s, r := range map[string]bool{
			"something.test": true,
			"def.abc.test":   true,
			"test":           false,
			".test":          true,
			"a.test":         true,
			"test.priv":      false,
		} {
			if r == true {
				It(fmt.Sprintf("sould match '%s'", s), func(s string) func() {
					return func() {
						Expect(pattern.Match(s)).To(BeTrue())
					}
				}(s))
			} else {
				It(fmt.Sprintf("sould not match '%s'", s), func(s string) func() {
					return func() {
						Expect(pattern.Match(s)).To(BeFalse())
					}
				}(s))
			}
		}
	})
})
