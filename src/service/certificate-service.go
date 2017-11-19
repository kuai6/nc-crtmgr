package service

import (
	"github.com/kuai6/nc-crtmgr/src/generator"
	"github.com/kuai6/nc-crtmgr/src/certificate"
	"time"
	"errors"
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
	var err error
	certificateDTO, err := c.generator.Generate(options)
	if err == nil {
		certificates := c.certificates.FindByGidAndDidAndStatus(options.Uid(), options.Did(), certificate.STATUS_ACTIVE)
		for _, crt_ := range certificates {
			crt_.SetNotActive()
			c.certificates.Store(crt_)
		}
	}
	crt := new(certificate.Certificate)
	crt.SetCreationDateTime(time.Now())
	crt.SetPrivateKey(certificateDTO.PrivateKey())
	crt.SetCertificate(certificateDTO.Certificate())
	crt.SetSerial(certificateDTO.Serial())
	crt.SetValidTill(certificateDTO.NotAfter())
	crt.SetActive()
	crt.SetUid(options.Uid())
	crt.SetDid(options.Did())
	if time.Now().After(certificateDTO.NotAfter()) {
		crt.SetNotActive()
	}
	err = c.certificates.Store(crt)
	return crt, err
}

func (c *CertificateService) ValidateCertificate(uid string, did string, candidate string) (bool, error) {
	// first of all try to get uid and did
	cUid, cDid, _ := c.generator.ParseUidDid(candidate)
	if cUid != "" && cDid != "" {
		if cUid != uid || cDid != did {
			return false, errors.New("Certificate UID or DID not match with given")
		}
	}

	//try to fetch intermediate certificate with give uid and did
	itrCrt := c.FetchActiveCertificateByUidAndDid(uid, did)
	if itrCrt == nil {
		return false, errors.New("Certificate wit given UID and DID not found")
	}

	return c.generator.Validate(candidate, itrCrt.GetCertificate())
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
