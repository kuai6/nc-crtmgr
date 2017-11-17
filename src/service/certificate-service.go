package service

import (
	"github.com/kuai6/nc-crtmgr/src/generator"
	"github.com/kuai6/nc-crtmgr/src/certificate"
	"time"
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

func (c *CertificateService) GenerateCertificate(options generator.Options) (*certificate.Certificate, error) {
	crt, err := c.generator.Generate(options)
	if err == nil {
		certificates := c.certificates.FindByGidAndDidAndStatus(options.Uid(), options.Did(), certificate.STATUS_ACTIVE)
		for _, crt := range certificates {
			crt.SetNotActive()
			c.certificates.Store(crt)
		}
	}
	c.Save(crt)
	return crt, err
}

func (c *CertificateService) ValidateCertificate(candidate string, parent *certificate.Certificate) (bool, error) {
	return c.generator.Validate(candidate, parent)
}

func (c *CertificateService) FetchCertificateObjectByItContent(candidate string) (*certificate.Certificate, error) {
	uid, did, err := c.generator.ParseUidDid(candidate)
	if err != nil {
		return nil, err
	}

	return c.FetchActiveCertificateByUidAndDid(uid, did), nil
}

func (c *CertificateService) FetchActiveCertificateByUidAndDid(uid string, did string) (*certificate.Certificate) {
	certificates := c.certificates.FindByGidAndDidAndStatus(uid, did, certificate.STATUS_ACTIVE)
	if len(certificates) > 0 {
		return certificates[0]
	}
	return nil
}

func (c *CertificateService) Withdraw(certificate *certificate.Certificate) error {
	certificate.SetWithdrawn()
	certificate.SetWithdrawalDateTime(time.Now())
	c.Save(certificate)
	return nil
}


func (c *CertificateService) RemoveExpired() {
	certificates := c.certificates.FindExpired()
	for _, crt := range certificates {
		crt.SetNotActive()
		c.certificates.Store(crt)
	}
}
