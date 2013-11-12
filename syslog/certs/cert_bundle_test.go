package certs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func TestCerts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cert Suite")
}

var _ = Describe("cert_bundle", func() {
	var (
		bundle CertBundle
	)
	BeforeEach(func() {
		bundle = NewCertBundle()
	})
	Describe("ImportFromFile", func() {
		Context("Valid file", func() {
			It("import a file", func() {
				err := bundle.ImportFromFile("./test/cert.pem")
				Expect(err).To(BeNil())
				Expect(len(bundle.CertPool.Subjects())).To(Equal(1))
			})
		})

		Context("Invalid file", func() {
			It("import a file", func() {
				err := bundle.ImportFromFile("./test/cert-does-not-exist.pem")
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
				Expect(len(bundle.CertPool.Subjects())).To(Equal(0))
			})
		})

	})
	Describe("ImportFromFiles", func() {
		Context("Valid file", func() {
			It("import a file", func() {
				err := bundle.ImportFromFiles([]string{"./test/cert.pem", "./test/cert-2.pem"})
				Expect(err).To(BeNil())
				Expect(len(bundle.CertPool.Subjects())).To(Equal(2))
			})
		})

		Context("Invalid file", func() {
			It("import a file", func() {
				err := bundle.ImportFromFiles([]string{"./test/cert-does-not-exist.pem", "./test/cert-does-not-exist.pem"})
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
				Expect(len(bundle.CertPool.Subjects())).To(Equal(0))
			})
		})
	})

	Describe("ImportBytes", func() {
		It("imports bytes", func() {
			data, err := ioutil.ReadFile("./test/cert.pem")
			Expect(err).To(BeNil())
			bundle.ImportBytes(data)
			Expect(len(bundle.CertPool.Subjects())).To(Equal(1))
		})
	})
})
