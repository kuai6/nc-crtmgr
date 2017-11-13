package generator

import (
	"github.com/kuai6/nc-crtmgr/src/certificate"
	"crypto/rsa"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"crypto/rand"
	"errors"
	"encoding/pem"
	"crypto/x509"
	"math/big"
	"crypto/x509/pkix"
	"time"
)

type CryptoTLS struct {
}

func (g CryptoTLS) Generate(options Options) (*certificate.Certificate, error) {
	var privateKey interface{}
	var publicKey interface{}
	var err error

	getPublicKey := func(priv interface{}) interface{} {
		switch k := priv.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey
		case *ecdsa.PrivateKey:
			return &k.PublicKey
		default:
			return nil
		}
	}

	blockForKey := func(priv interface{}) (*pem.Block, error) {
		switch k := priv.(type) {
		case *rsa.PrivateKey:
			return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
		case *ecdsa.PrivateKey:
			b, err := x509.MarshalECPrivateKey(k)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Unable to marshal ECDSA private key: %s", err))
			}
			return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil

		default:
			return nil, nil
		}
	}

	genSerial := func() (*big.Int, error) {
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		return rand.Int(rand.Reader, serialNumberLimit)
	}

	switch options.EcdsaCurve() {
	case "":
		privateKey, err = rsa.GenerateKey(rand.Reader, options.RsaBits())
	case "P224":
		privateKey, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		privateKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, errors.New(fmt.Sprintf("Unrecognized elliptic curve: %s", options.EcdsaCurve()))
	}
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to generate private key: %s", err))
	}

	publicKey = getPublicKey(privateKey)

	serialNumber := new(big.Int)
	serialNumber, err = genSerial()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to generate serial number: %s", err))
	}

	var notBefore time.Time
	if len(options.ValidFrom()) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse(time.RFC3339, options.ValidFrom())
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to parse creation date: %s", err))
		}
	}

	notAfter := time.Now().Add(time.Duration(options.DefaultTTL()) * 24 * time.Hour)
	if len(options.ValidFor()) != 0 {
		notAfter, err = time.Parse(time.RFC3339, options.ValidFor())
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to parse expiration date: %s", err))
		}
	}

	cert := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		IsCA:      true,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	prk := new(pem.Block)
	prk, err = blockForKey(privateKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to generate block: %s", err))
	}

	ck, err := x509.CreateCertificate(rand.Reader, &cert, &cert, publicKey, privateKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to generate certificate: %s", err))
	}

	c := new(certificate.Certificate)
	c.CreationDateTime = time.Now()
	c.PrivateKey = fmt.Sprintf("%s", pem.EncodeToMemory(prk))
	c.Certificate = fmt.Sprintf("%s", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ck}))
	c.Serial = serialNumber.String()
	c.ValidTill = notAfter
	c.SetActive()
	if time.Now().After(c.ValidTill) {
		c.SetNotActive()
	}
	return c, nil
}

func (g CryptoTLS) validate(certificate certificate.Certificate) bool {
	return true
}
