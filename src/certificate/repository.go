package certificate

import "math/big"

type Repository interface {
	Store(certificate *Certificate) error
	Find(serial big.Int) (*Certificate, error)
	FindAll() []*Certificate
	FindExpired() []*Certificate
	FindByGidAndDidAndStatus(gid string, did string, status int) []*Certificate
}
