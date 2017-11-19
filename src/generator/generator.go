package generator

import (
	"time"
)

type Subject struct {
	CommonName         string
	Country            string
	Province           string
	Locality           string
	Organization       string
	OrganizationalUnit string
}

type Generator interface {
	Generate(options Options) (*CertificateDTO, error)
	Validate(content string, intermediate string) (bool, error)
	ParseUidDid(content string) (string, string, error)
	ParseDates(content string) (*time.Time, *time.Time, error)
}

type CertificateDTO struct {
	certificate string
	privateKey  string
	notAfter    time.Time
	notBefore   time.Time
	serial      string
}

func (cd CertificateDTO) Certificate() string {
	return cd.certificate
}

func (cd CertificateDTO) PrivateKey() string {
	return cd.privateKey
}

func (cd CertificateDTO) NotAfter() time.Time {
	return cd.notAfter
}

func (cd CertificateDTO) NotBefore() time.Time {
	return cd.notBefore
}

func (cd CertificateDTO) Serial() string {
	return cd.serial
}