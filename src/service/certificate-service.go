package service

import (
	"github.com/kuai6/nc-crtmgr/src/generator"
	"github.com/kuai6/nc-crtmgr/src/certificate"
)

type CertificateServiceInterface interface {
	Save(certificate *certificate.Certificate) error
	Generate(options generator.Options) error
	RemoveExpired()
}

type CertificateService struct {
	certificates certificate.Repository
	generator    generator.Generator
}

func NewCertificateService(repository certificate.Repository, generator generator.Generator) *CertificateService {
	return &CertificateService{
		certificates: repository,
		generator:    generator,
	}
}

func (c *CertificateService) Save(certificate *certificate.Certificate) error {
	return c.certificates.Store(certificate)
}

func (s *CertificateService) GenerateCertificate(options generator.Options) (*certificate.Certificate, error) {
	return s.generator.Generate(options)
}

func (c *CertificateService) RemoveExpired() {
	certificates := c.certificates.FindExpired()
	for _, crt := range certificates {
		crt.SetNotActive()
		c.certificates.Store(crt)
	}
}
