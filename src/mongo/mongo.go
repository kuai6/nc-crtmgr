package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/kuai6/nc-crtmgr/src/certificate"
	"math/big"
	"errors"
	"fmt"
	"time"
)

type CertificateRepository struct {
	collectionName string
	db             string
	session        *mgo.Session
}

func NewCertificateRepository(db string, session *mgo.Session) (certificate.Repository, error) {
	r := &CertificateRepository{
		collectionName: "certificate",
		db:             db,
		session:        session,
	}

	index := mgo.Index{
		Key:        []string{"serial"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(r.collectionName)

	if err := c.EnsureIndex(index); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *CertificateRepository) Store(certificate *certificate.Certificate) error {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(r.collectionName)

	_, err := c.Upsert(bson.M{"serial": certificate.GetSerial()}, bson.M{"$set": certificate})

	return err
}

func (r *CertificateRepository) Find(serial big.Int) (*certificate.Certificate, error) {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(r.collectionName)

	var result certificate.Certificate
	if err := c.Find(bson.M{"serial": serial}).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.New(fmt.Sprintf("certificate with serial %d not found", serial))
		}
		return nil, err
	}

	return &result, nil
}

func (r *CertificateRepository) FindAll() []*certificate.Certificate {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(r.collectionName)

	var result []*certificate.Certificate
	if err := c.Find(bson.M{}).All(&result); err != nil {
		return []*certificate.Certificate{}
	}

	return result
}

func (r *CertificateRepository) FindExpired() []*certificate.Certificate {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(r.collectionName)

	var result []*certificate.Certificate
	if err := c.Find(bson.M{"validtill": bson.M{"$lte": time.Now()}}).All(&result); err != nil {
		return []*certificate.Certificate{}
	}

	return result
}

func (r *CertificateRepository) FindByGidAndDidAndStatus(uid string, did string, status int) []*certificate.Certificate {
	sess := r.session.Copy()
	defer sess.Close()

	c := sess.DB(r.db).C(r.collectionName)

	var result []*certificate.Certificate
	if err := c.Find(bson.M{"$and": []bson.M{
		{"uid": uid},
		{"did": did},
		{"status": status},
	}}).All(&result); err != nil {
		return []*certificate.Certificate{}
	}

	return result
}
