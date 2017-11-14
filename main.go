package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"

	"github.com/julienschmidt/httprouter"
	"github.com/kuai6/nc-crtmgr/src/mongo"
	"github.com/kuai6/nc-crtmgr/src/generator"
	"github.com/kuai6/nc-crtmgr/src/certificate"
	"github.com/kuai6/nc-crtmgr/src/service"
	"github.com/mileusna/crontab"
	"github.com/sarulabs/di"
	"gopkg.in/mgo.v2"
)

type GenerateRequest struct {
	Uid      string `json:"uid"`
	Did      string `json:"did"`
	Password string `json:"password"`
	ValidFrom string `json:"valid_from"`
	ValidFor  string `json:"valid_for"`
}

type GenerateResponse struct {
	Uid         string `json:"uid"`
	Did         string `json:"did"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
	ValidTill   string `json:"valid_till"`
}

type ValidateRequest struct {
}

type ValidateResponse struct {
}

type WithdrawalRequest struct {
}

type WithdrawalResponse struct {
}

var context di.Context

func main() {
	//@TODO move to context
	InitLogger(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	builder, _ := di.NewBuilder()
	builder.AddDefinition(di.Definition{
		Name:  "config",
		Scope: di.App,
		Build: func(ctx di.Context) (interface{}, error) {
			return GetConfig(), nil
		},
	})

	builder.AddDefinition(di.Definition{
		Name:  "router",
		Scope: di.App,
		Build: func(ctx di.Context) (interface{}, error) {
			return InitRouter(), nil
		},
	})

	builder.AddDefinition(di.Definition{
		Name:  "mongo",
		Scope: di.App,
		Build: func(ctx di.Context) (interface{}, error) {
			config := ctx.Get("config").(*Config)
			mongoDsn := fmt.Sprintf("mongodb://%s:%d/%s", config.DbConfig.Host, config.DbConfig.Port, config.DbConfig.Name)
			session, err := mgo.Dial(mongoDsn)
			if err != nil {
				Error.Fatal(err)
			}
			//defer session.Close()
			session.SetMode(mgo.Monotonic, true)
			return session, nil
		},
	})

	builder.AddDefinition(di.Definition{
		Name:  "generator",
		Scope: di.App,
		Build: func(ctx di.Context) (interface{}, error) {
			config := ctx.Get("config").(*Config)
			g := new(generator.CryptoTLS)
			g.DefaultSubject = generator.Subject{
				CommonName:         config.CertificateSubject.CommonName,
				Country:            config.CertificateSubject.Country,
				Province:           config.CertificateSubject.Province,
				Locality:           config.CertificateSubject.Locality,
				Organization:       config.CertificateSubject.Organization,
				OrganizationalUnit: config.CertificateSubject.OrganizationalUnit,
			}

			crt, err := ioutil.ReadFile(config.RootCertPath)
			if err != nil {
				Error.Fatal(fmt.Sprintf("Cant't read root cerificate %s", config.RootCertPath))
			}
			key, err := ioutil.ReadFile(config.RootCertKeyPath)
			if err != nil {
				Error.Fatal(fmt.Sprintf("Cant't read root cerificate private key %s", config.RootCertKeyPath))
			}
			g.LoadRootCA(crt, key)
			g.DefaultTTL = config.CertTTL
			g.RsaBits = config.KeyRSABits
			return g, nil
		},
	})
	context = builder.Build()

	router := context.Get("router").(*httprouter.Router)
	config := context.Get("config").(*Config)

	cron := crontab.New()
	cron.AddJob("* * * * *", CleanUp)

	err := http.ListenAndServeTLS(
		fmt.Sprintf("%s:%d", config.HttpConfig.Listen, config.HttpConfig.Port),
		config.HttpConfig.SSLCertPath,
		config.HttpConfig.SSLCertKeyPath,
		router)
	if err != nil {
		Error.Fatal("ListenAndServe: ", err)
	}
}

func GenerateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	var gr GenerateRequest
	err := decoder.Decode(&gr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad request!"))
	}
	defer r.Body.Close()

	done := make(chan certificate.Certificate)
	e := make(chan error)
	go func() {
		config := context.Get("config").(*Config)
		session := context.Get("mongo").(*mgo.Session)
		gen := context.Get("generator").(generator.Generator)

		repository, _ := mongo.NewCertificateRepository(config.DbConfig.Name, session)
		certificateService := service.NewCertificateService(repository, gen)

		o := generator.Options{}
		o.SetValidFrom(gr.ValidFrom)
		o.SetValidFor(gr.ValidFor)
		o.SetPassword(gr.Password)

		c, err := certificateService.GenerateCertificate(o)
		if err != nil {
			Error.Println(err)
			e <- err
			close(done)
			close(e)
			return
		}

		c.Did = gr.Did
		c.Uid = gr.Uid
		certificateService.Save(c)
		done <- *c
		close(e)
		close(done)
	}()

	select {
	case err := <-e:
		generateResult := ErrorResponse{err.Error()}
		result, _ := json.Marshal(generateResult)
		w.Write(result)
		w.WriteHeader(http.StatusInternalServerError)
	case c := <-done:
		generateResult := GenerateResponse{}
		generateResult.Uid = c.Uid
		generateResult.Did = c.Did
		generateResult.Certificate = c.Certificate
		generateResult.PrivateKey = c.PrivateKey
		generateResult.ValidTill = c.ValidTill.String()
		result, err := json.Marshal(generateResult)
		if err != nil {
			msg := fmt.Sprintf("Internal Server Error: %s", err)
			Error.Fatal(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(result)
	}
}

func ValidateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func CleanUp() {
	go func() {
		config := context.Get("config").(*Config)
		session := context.Get("mongo").(*mgo.Session)
		gen := context.Get("generator").(generator.Generator)

		repository, _ := mongo.NewCertificateRepository(config.DbConfig.Name, session)
		certificateService := service.NewCertificateService(repository, gen)
		certificateService.RemoveExpired()
	}()
}
