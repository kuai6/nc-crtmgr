package generator

import "github.com/kuai6/nc-crtmgr/src/certificate"

type Generator interface {
	Generate(options Options) (*certificate.Certificate, error)
}
