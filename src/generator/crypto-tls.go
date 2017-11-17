package generator

import (
	"github.com/kuai6/nc-crtmgr/src/certificate"
	"crypto/rsa"
	"fmt"
	"crypto/rand"
	"errors"
	"crypto/x509"
	"math/big"
	"crypto/x509/pkix"
	"encoding/pem"
	"time"
	"encoding/asn1"
	"encoding/base64"
)

type CryptoTLS struct {
	DefaultSubject Subject
	RsaBits        int
	DefaultTTL     int
	rootCACrt      *x509.Certificate
	rootCAKey      *rsa.PrivateKey
}

func (g *CryptoTLS) LoadRootCA(crt []byte, key [] byte) error {
	var err error
	bcrt, _ := pem.Decode(crt)
	if g.rootCACrt, err = x509.ParseCertificate(bcrt.Bytes); err != nil {
		return errors.New(fmt.Sprintf("Failed to parse root certificate: %s", err.Error()))
	}
	bkey, _ := pem.Decode(key)
	if g.rootCAKey, err = x509.ParsePKCS1PrivateKey(bkey.Bytes); err != nil {
		return errors.New(fmt.Sprintf("Failed to parse root certificate private key: %s", err.Error()))
	}
	return nil
}

func (g *CryptoTLS) Generate(options Options) (*certificate.Certificate, error) {
	var err error
	// First of all gen new private key
	newCrtPrivateKey, _ := rsa.GenerateKey(rand.Reader, g.RsaBits)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to generate private key: %s", err))
	}

	//then gen certificate serial
	genSerial := func() (*big.Int, error) {
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		return rand.Int(rand.Reader, serialNumberLimit)
	}
	serialNumber := new(big.Int)
	serialNumber, err = genSerial()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to generate serial number: %s", err))
	}

	// now we need to create certificate request
	tcsr := x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{g.DefaultSubject.Country},
			Organization:       []string{g.DefaultSubject.Organization},
			OrganizationalUnit: []string{g.DefaultSubject.OrganizationalUnit},
			Locality:           []string{},
			Province:           []string{},
			SerialNumber:       fmt.Sprintf("%s", serialNumber),
			CommonName:         g.DefaultSubject.CommonName,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	bcsr, err := x509.CreateCertificateRequest(rand.Reader, &tcsr, newCrtPrivateKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to generate CSR: %s", err))
	}
	csr, _ := x509.ParseCertificateRequest(bcsr)

	// resolve certificate dates
	var notBefore time.Time
	if len(options.ValidFrom()) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse(time.RFC3339, options.ValidFrom())
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to parse creation date: %s", err))
		}
	}

	notAfter := time.Now().Add(time.Duration(g.DefaultTTL) * 24 * time.Hour)
	if len(options.ValidFor()) != 0 {
		notAfter, err = time.Parse(time.RFC3339, options.ValidFor())
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to parse expiration date: %s", err))
		}
	}
	//Custom OID's
	uid := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 2}
	did := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 2}

	// generate certificate with sign
	cert := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,
		Extensions:			csr.Extensions,
		ExtraExtensions:    []pkix.Extension{
			{Id: uid, Value: []byte(fmt.Sprintf("UID:%s", options.Uid()))},
			{Id: did, Value: []byte(fmt.Sprintf("DID:%s", options.Did()))},
		},
		SerialNumber: serialNumber,
		Issuer:       g.rootCACrt.Subject,
		Subject:      csr.Subject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		IsCA: 		  true,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	ck, err := x509.CreateCertificate(rand.Reader, &cert, g.rootCACrt, csr.PublicKey, g.rootCAKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to generate certificate: %s", err))
	}

	var pkey, crt *pem.Block
	if options.Password() != "" {
		pkey, err = x509.EncryptPEMBlock(
			rand.Reader, "RSA PRIVATE KEY",
			x509.MarshalPKCS1PrivateKey(newCrtPrivateKey), []byte(options.Password()), x509.PEMCipherAES128)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to enctypt certificate key with password: %s", err))
		}
	} else {
		pkey = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(newCrtPrivateKey)}
	}

	crt = &pem.Block{Type: "CERTIFICATE", Bytes: ck}

	c := new(certificate.Certificate)
	c.CreationDateTime = time.Now()
	c.PrivateKey = base64.StdEncoding.EncodeToString([]byte(pem.EncodeToMemory(pkey)))
	c.Certificate = base64.StdEncoding.EncodeToString([]byte(pem.EncodeToMemory(crt)))
	c.Serial = serialNumber.String()
	c.ValidTill = notAfter
	c.SetActive()
	if time.Now().After(c.ValidTill) {
		c.SetNotActive()
	}
	return c, nil
}

func (g *CryptoTLS) Validate(content string, parent *certificate.Certificate) (bool, error) {

	opts := x509.VerifyOptions{
		Roots: x509.NewCertPool(),
		Intermediates: x509.NewCertPool(),

		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	opts.CurrentTime = time.Now()
	opts.Roots.AddCert(g.rootCACrt)

	if parent != nil {
		content, _ := base64.StdEncoding.DecodeString(parent.Certificate)
		bcrt, _ := pem.Decode(content)
		var pcrt *x509.Certificate
		var err error
		if pcrt, err = x509.ParseCertificate(bcrt.Bytes); err != nil {
			return false, errors.New(fmt.Sprintf("Failed to parse certificate: %s", err.Error()))
		}
		opts.Intermediates.AddCert(pcrt)
	}

	var crt *x509.Certificate
	var err error
	bcrt, _ := pem.Decode([]byte(content))
	if crt, err = x509.ParseCertificate(bcrt.Bytes); err != nil {
		return false, errors.New(fmt.Sprintf("Failed to parse certificate: %s", err.Error()))
	}

	_ , err = crt.Verify(opts)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Failed to validate certificate: %s", err.Error()))
	}

	return true, nil
}
