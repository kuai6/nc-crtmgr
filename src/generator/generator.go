package generator

import "github.com/kuai6/nc-crtmgr/src/certificate"

type Subject struct {
	CommonName         string
	Country            string
	Province           string
	Locality           string
	Organization       string
	OrganizationalUnit string
}

type Generator interface {
	Generate(options Options) (*certificate.Certificate, error)
}
