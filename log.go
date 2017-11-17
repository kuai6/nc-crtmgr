package main

import (
	"log"
	"github.com/op/go-logging"
	"os"
	"flag"
)

var (
	logger  *logging.Logger
	verbose = flag.Bool("v", false, "Verbose")
)

func InitLogger() {
	logger = logging.MustGetLogger("app")

	var l = "ERROR"
	if *verbose == true {
		l = "DEBUG"
	}

	level, err := logging.LogLevel(l)
	if err != nil {
		log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime).Panicln(err.Error())
		os.Exit(1)
	}

	backend := logging.NewLogBackend(os.Stdout, "", 0)

	backendFormatted := logging.NewBackendFormatter(
		backend,
		logging.MustStringFormatter("%{time:2006-01-02 15:04:05} [%{level:.4s}]: %{message}"))

	backendLeveled := logging.AddModuleLevel(backendFormatted)
	backendLeveled.SetLevel(level, "")

	logging.SetBackend(backendLeveled);
}
