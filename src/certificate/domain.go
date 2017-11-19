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
	Uid         string
	Did         string
	PrivateKey  string
	Certificate string
	Serial      string

	CreationDateTime   time.Time
	ValidTill          time.Time
	WithdrawalDateTime time.Time

	Status int
}

func (c *Certificate) SetDid(value string) {
	c.Did = value
}

func (c Certificate) GetDid() string {
	return c.Did
}

func (c *Certificate) SetUid(value string) {
	c.Uid = value
}

func (c Certificate) GetUid() string {
	return c.Uid
}

func (c *Certificate) SetPrivateKey(value string) {
	c.PrivateKey = base64.StdEncoding.EncodeToString([]byte(value))
}

func (c Certificate) GetPrivateKey() string {
	s, _ := base64.StdEncoding.DecodeString(c.PrivateKey)
	return string(s)
}

func (c Certificate) GetPrivateKeyBase64() string {
	return c.PrivateKey
}

func (c *Certificate) SetCertificate(value string) {
	c.Certificate = base64.StdEncoding.EncodeToString([]byte(value))
}

func (c Certificate) GetCertificate() string {
	s, _ := base64.StdEncoding.DecodeString(c.Certificate)
	return string(s)
}

func (c Certificate) GetCertificateBase64() string {
	return c.Certificate
}

func (c *Certificate) SetSerial(value string) {
	c.Serial = value
}

func (c Certificate) GetSerial() string {
	return c.Serial
}

func (c *Certificate) SetCreationDateTime(value time.Time) {
	c.CreationDateTime = value
}

func (c Certificate) GetCreationDateTime() time.Time {
	return c.CreationDateTime
}

func (c *Certificate) SetValidTill(value time.Time) {
	c.ValidTill = value
}

func (c Certificate) GetValidTill() time.Time {
	return c.ValidTill
}

func (c *Certificate) SetWithdrawalDateTime(value time.Time) {
	c.WithdrawalDateTime = value
}

func (c Certificate) GetWithdrawalDateTime() time.Time {
	return c.WithdrawalDateTime
}

func (c Certificate) GetStatus() int {
	return c.Status
}

func (c *Certificate) SetActive() {
	c.Status = STATUS_ACTIVE
}

func (c *Certificate) SetWithdrawn() {
	c.Status = STATUS_WITHDRAWN
}

func (c *Certificate) SetNotActive() {
	c.Status = STATUS_NOT_ACTIVE
}
