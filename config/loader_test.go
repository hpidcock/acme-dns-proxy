package config_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/matthiasng/dns-challenge-proxy/config"

	"github.com/spf13/afero"
)

var _ = Describe("Loader", func() {
	Describe("load file if the file does not exist", func() {
		It("should fail", func() {
			_, err := NewFileLoader(afero.NewMemMapFs(), "unknow.yml").Load()
			Expect(err).To(MatchError(ContainSubstring("file does not exist")))
		})
	})

	Describe("load file", func() {
		var (
			fs afero.Fs

			cfg *Config
			err error

			withContent = func(c string) {
				BeforeEach(func() {
					Expect(afero.WriteFile(fs, "config.yml", []byte(c), 0644)).To(Succeed())
				})
			}
			withEnvVar = func(key, value string) {
				BeforeEach(func() {
					os.Setenv(key, value)
				})
				AfterEach(func() {
					os.Unsetenv(key)
				})
			}
		)

		BeforeEach(func() {
			fs = afero.NewMemMapFs()
		})

		JustBeforeEach(func() {
			cfg, err = NewFileLoader(fs, "config.yml").Load()
		})

		Context("with an empty file", func() {
			withContent("")

			It("should return nil", func() {
				Expect(cfg).To(BeNil())
			})
			It("should fail", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("file is empty")))
			})
		})

		Context("with invalid yaml syntax", func() {
			withContent("invalid-yaml-content")

			It("should return nil", func() {
				Expect(cfg).To(BeNil())
			})
			It("should fail", func() {
				Expect(err).To(MatchError(ContainSubstring("cannot unmarshal")))
			})
		})

		Context("with a valid file", func() {
			withContent(`
server:
  addr: ":8080"

provider:
  type: digitalocean
  variables:
    DO_TTL: 120
    DO_AUTH_TOKEN: "SOME_TOKEN"
    
access_rules:
  - pattern: "sub.test.local"
    token: b5aa632dc6035e426152b50f80f654ad27b329107267d445bd30c673c90f82be
  - pattern: "*test.local"
    token: 76734be06e3eaa967fa82746bac47e9621f291a5a18222d32016b7febacd4548
  - pattern: "private.test.local"
    token: f221311c7ec8b6e88261e23bf62b5c6d7915e9c45f7b6f6f6be2958084aeb6ba
`)

			It("should succeed", func() {
				Expect(err).To(Succeed())
			})

			Specify("server fields", func() {
				Expect(cfg.Server.Addr).Should(Equal(":8080"))
			})

			Specify("provider fields", func() {
				Expect(cfg.Provider.Type).Should(Equal("digitalocean"))
				Expect(cfg.Provider.Variables).Should(Equal(map[string]string{
					"DO_TTL":        "120",
					"DO_AUTH_TOKEN": "SOME_TOKEN",
				}))
			})

			Specify("access_rules fields", func() {
				Expect(cfg.AccessRules).Should(Equal(AccessRules{
					{Pattern: "sub.test.local", Token: "b5aa632dc6035e426152b50f80f654ad27b329107267d445bd30c673c90f82be"},
					{Pattern: "*test.local", Token: "76734be06e3eaa967fa82746bac47e9621f291a5a18222d32016b7febacd4548"},
					{Pattern: "private.test.local", Token: "f221311c7ec8b6e88261e23bf62b5c6d7915e9c45f7b6f6f6be2958084aeb6ba"},
				}))
			})
		})

		Context("with invalid template syntax", func() {
			withContent("{{ invalid }}")

			It("should return nil", func() {
				Expect(cfg).To(BeNil())
			})
			It("sould faild", func() {
				Expect(err).To(MatchError(ContainSubstring(`"invalid" not defined`)))
			})
		})

		Context("with unknown template variable", func() {
			withContent("{{ .unknown }}")

			It("should return nil", func() {
				Expect(cfg).To(BeNil())
			})
			It("sould faild", func() {
				Expect(err).To(MatchError(ContainSubstring("can't evaluate field unknown")))
			})
		})

		Context("with valid template syntax", func() {
			addr := "localhost:8080"
			withEnvVar("SERVER_ADDR", addr)
			withContent(`
server:
  addr: {{ .Env.SERVER_ADDR }}
`)

			It("should succeed", func() {
				Expect(err).To(Succeed())
			})
			Specify("server.addr == env var", func() {
				Expect(cfg.Server.Addr).Should(Equal(addr))
			})
		})
	})
})

// package config_test

// import (
// 	"os"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	. "github.com/matthiasng/dns-challenge-proxy/config"

// 	"github.com/spf13/afero"
// )

// var _ = Describe("Loader", func() {
// 	var (
// 		content string
// 		fs afero.Fs

// 		cfg *Config
// 		err error

// 		withContent = func(c string) {
// 			BeforeEach(func() {
// 				content = c
// 			})
// 		}
// 		withEnvVar = func(key, value string) {
// 			BeforeEach(func() {
// 				os.Setenv(key, value)
// 			})
// 			AfterEach(func() {
// 				os.Unsetenv(key)
// 			})
// 		}
// 	)

// 	BeforeEach(func() {
// 		fs = afero.NewMemMapFs()
// 	})

// 	JustBeforeEach(func() {
// 		Expect(afero.WriteFile(fs, "config.yml", []byte(content), 0644)).To(Succeed())
// 		cfg, err = NewFileLoader(fs, "config.yml").Load()
// 	})

// 	Describe("loading file", func() {
// 		Context("when the file do not exists", func() {
// 			It("should fail", func() {
// 				_, err := NewFileLoader(afero.NewMemMapFs(), "unknow.yml").Load()
// 				Expect(err).To(MatchError(ContainSubstring("cannot find the file specified")))
// 			})
// 		})

// 		Context("with an empty file", func() {
// 			withContent("")

// 			It("should fail", func() {
// 				Expect(err).To(HaveOccurred())
// 				Expect(err).To(MatchError(ContainSubstring("file is empty")))
// 			})
// 		})

// 		Context("with invalid yaml syntax", func() {
// 			withContent("invalid-yaml-content")

// 			It("should fail", func() {
// 				Expect(err).To(MatchError(ContainSubstring("cannot unmarshal")))
// 			})
// 		})

// 		Context("with invalid template syntax", func() {
// 			withContent("{{ invalid }}")

// 			It("sould faild", func() {
// 				Expect(err).To(MatchError(ContainSubstring(`"invalid" not defined`)))
// 			})
// 		})

// 		Context("with unknown template variable", func() {
// 			withContent("{{ .unknown }}")

// 			It("sould faild", func() {
// 				Expect(err).To(MatchError(ContainSubstring("can't evaluate field unknown")))
// 			})
// 		})

// 		Context("with valid template syntax", func() {
// 			withEnvVar("SERVER_ADDR", "localhost:8080")
// 			withContent(`
// server:
//   addr: {{ .Env.SERVER_ADDR }}
// `)

// 			It("should succeed", func() {
// 				Expect(err).To(Succeed())
// 			})
// 			Specify("access_rules order is stable", func() {
// 				Expect(cfg.Server.Addr).Should(Equal("localhost:8080"))
// 			})
// 		})
// 	})
// })
