package certificate

import (
	"time"
	"encoding/base64"
)

const (
	STATUS_ACTIVE     = 1
	STATUS_WITHDRAWN  = 2
	STATUS_NOT_ACTIVE = 0
)

type Certificate struct {
	uid         string
	did         string
	privateKey  string
	certificate string
	serial      string

	creationDateTime   time.Time
	validTill          time.Time
	withdrawalDateTime time.Time

	status int
}

func (c *Certificate) SetDid(value string) {
	c.did = value
}

func (c Certificate) Did() string {
	return c.did
}

func (c *Certificate) SetUid(value string) {
	c.uid = value
}

func (c Certificate) Uid() string {
	return c.uid
}

func (c *Certificate) SetPrivateKey(value string) {
	c.privateKey = base64.StdEncoding.EncodeToString([]byte(value))
}

func (c Certificate) PrivateKey() string {
	s, _ := base64.StdEncoding.DecodeString(c.privateKey)
	return string(s)
}

func (c Certificate) PrivateKeyBase64() string {
	return c.privateKey
}

func (c *Certificate) SetCertificate(value string) {
	c.certificate = base64.StdEncoding.EncodeToString([]byte(value))
}

func (c Certificate) Certificate() string {
	s, _ := base64.StdEncoding.DecodeString(c.certificate)
	return string(s)
}

func (c Certificate) CertificateBase64() string {
	return c.certificate
}

func (c *Certificate) SetSerial(value string) {
	c.serial = value
}

func (c Certificate) Serial() string {
	return c.serial
}

func (c *Certificate) SetCreationDateTime(value time.Time) {
	c.creationDateTime = value
}

func (c Certificate) CreationDateTime() time.Time {
	return c.creationDateTime
}

func (c *Certificate) SetValidTill(value time.Time) {
	c.validTill = value
}

func (c Certificate) ValidTill() time.Time {
	return c.validTill
}

func (c *Certificate) SetWithdrawalDateTime(value time.Time) {
	c.withdrawalDateTime = value
}

func (c Certificate) WithdrawalDateTime() time.Time {
	return c.withdrawalDateTime
}

func (c Certificate) Status() int {
	return c.status
}

func (c *Certificate) SetActive() {
	c.status = STATUS_ACTIVE
}

func (c *Certificate) SetWithdrawn() {
	c.status = STATUS_WITHDRAWN
}

func (c *Certificate) SetNotActive() {
	c.status = STATUS_NOT_ACTIVE
}
