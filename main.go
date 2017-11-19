package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"fmt"

	"github.com/julienschmidt/httprouter"
	"github.com/kuai6/nc-crtmgr/src/mongo"
	"github.com/kuai6/nc-crtmgr/src/generator"
	"github.com/kuai6/nc-crtmgr/src/service"
	"github.com/mileusna/crontab"
	"github.com/sarulabs/di"
	"gopkg.in/mgo.v2"
	"flag"
	"encoding/base64"
	"time"
)

var (
	cliConfigFilePath = flag.String("config", "", "Config file path")
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
	Result      bool   `json:"result"`
	Reason      string `json:"reason"`
}

type ValidateRequest struct {
	Uid         string `json:"uid"`
	Did         string `json:"did"`
	Certificate string `json:"certificate"`
}

type ValidateResponse struct {
	Uid         string `json:"uid"`
	Did         string `json:"did"`
	Result      bool   `json:"result"`
	Reason		string `json:"reason"`
}

type WithdrawalRequest struct {
	Uid         string `json:"uid"`
	Did         string `json:"did"`
	Certificate string `json:"certificate"`
}

type WithdrawalResponse struct {
	Uid    string `json:"uid"`
	Did    string `json:"did"`
	Result bool   `json:"result"`
	Reason string `json:"reason"`
}

var context di.Context

func main() {
	flag.Parse()
	InitLogger()

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
				logger.Critical(err)
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
				logger.Criticalf("Cant't read root cerificate %s", config.RootCertPath)
			}
			key, err := ioutil.ReadFile(config.RootCertKeyPath)
			if err != nil {
				logger.Criticalf("Cant't read root cerificate private key %s", config.RootCertKeyPath)
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
		logger.Fatal(err)
	}
}

func GenerateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	var result []byte
	var err error
	var gr GenerateRequest

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&gr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad request!"))
	}
	defer r.Body.Close()

	done := make(chan GenerateResponse)
	go func() {
		var response GenerateResponse
		response.Uid = gr.Uid
		response.Did = gr.Did
		response.Result = true

		config := context.Get("config").(*Config)
		session := context.Get("mongo").(*mgo.Session)
		gen := context.Get("generator").(generator.Generator)

		repository, _ := mongo.NewCertificateRepository(config.DbConfig.Name, session)
		certificateService := service.NewCertificateService(repository, gen)

		o := generator.Options{}
		o.SetValidFrom(gr.ValidFrom)
		o.SetValidFor(gr.ValidFor)
		o.SetPassword(gr.Password)
		o.SetUid(gr.Uid)
		o.SetDid(gr.Did)

		c, err := certificateService.GenerateCertificate(o)
		if err != nil {
			response.Result = false
			response.Reason = err.Error()
			done <- response
			close(done)
			return
		}

		response.Certificate = c.GetCertificateBase64()
		response.PrivateKey = c.GetPrivateKeyBase64()
		response.ValidTill = c.GetValidTill().Format(time.RFC3339)
		done <- response
		close(done)
	}()

	result, err = json.Marshal(<-done)
	if err != nil {
		msg := fmt.Sprintf("Internal Server Error: %s", err)
		logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(result)
}

func ValidateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	var result []byte
	var err error
	var vr ValidateRequest

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&vr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad request!"))
	}
	defer r.Body.Close()

	done := make(chan ValidateResponse)
	go func() {
		var response ValidateResponse
		response.Uid = vr.Uid
		response.Did = vr.Did
		response.Result = true

		config := context.Get("config").(*Config)
		session := context.Get("mongo").(*mgo.Session)
		gen := context.Get("generator").(generator.Generator)

		repository, _ := mongo.NewCertificateRepository(config.DbConfig.Name, session)
		certificateService := service.NewCertificateService(repository, gen)

		sDec, err := base64.StdEncoding.DecodeString(vr.Certificate)
		if err != nil {
			response.Result = false
			response.Reason = err.Error()
			done <- response
			close(done)
			return
		}

		response.Result, err = certificateService.ValidateCertificate(vr.Uid, vr.Did, fmt.Sprintf("%s", sDec))
		if err != nil {
			response.Reason = err.Error()
		}

		done <- response
		close(done)
	}()

	result, err = json.Marshal(<-done)
	if err != nil {
		msg := fmt.Sprintf("Internal Server Error: %s", err)
		logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(result)
}

func WithdrawalHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	var result []byte
	var err error
	var wr WithdrawalRequest

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&wr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad request!"))
		return
	}
	defer r.Body.Close()

	done := make(chan WithdrawalResponse)
	go func() {
		var response WithdrawalResponse
		response.Uid = wr.Did
		response.Did = wr.Did
		response.Result = true

		config := context.Get("config").(*Config)
		session := context.Get("mongo").(*mgo.Session)
		gen := context.Get("generator").(generator.Generator)

		repository, _ := mongo.NewCertificateRepository(config.DbConfig.Name, session)
		certificateService := service.NewCertificateService(repository, gen)

		sDec, err := base64.StdEncoding.DecodeString(wr.Certificate)
		if err != nil {
			response.Result = false
			response.Reason = err.Error()
			done <- response
			close(done)
			return
		}

		_, err = certificateService.ValidateCertificate(wr.Uid, wr.Did, fmt.Sprintf("%s", sDec))
		if err != nil {
			response.Result = false
			response.Reason = err.Error()
			done <- response
			close(done)
			return
		}

		cert, err := certificateService.FetchCertificateObjectByItContent(fmt.Sprintf("%s", sDec))
		if err != nil {
			response.Result = false
			response.Reason = err.Error()
			done <- response
			close(done)
			return
		}
		if cert == nil {
			response.Result = false
			response.Reason = "Certificate not found"
			done <- response
			close(done)
			return
		}

		err = certificateService.Withdraw(cert)
		if err != nil {
			response.Result = false
			response.Reason = err.Error()
			done <- response
			close(done)
			return
		}

		done <- response
		close(done)
	}()

	result, err = json.Marshal(<-done)
	if err != nil {
		msg := fmt.Sprintf("Internal Server Error: %s", err)
		logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(result)
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
