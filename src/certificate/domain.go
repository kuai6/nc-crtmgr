package certificate

import "time"

const (
	STATUS_ACTIVE     = 1
	STATUS_WITHDRAWN  = 2
	STATUS_NOT_ACTIVE = 0
)

type Certificate struct {
	Uid         string `json:"uid"`
	Did         string `json:"did"`
	PrivateKey  string `json:"private_key"`
	Certificate string `json:"certificate"`
	Serial      string `json:"serial"`

	CreationDateTime time.Time `json:"creation_date_time"`
	ValidTill        time.Time `json:"valid_till"`

	Status int
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
