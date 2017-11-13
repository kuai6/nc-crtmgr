package main

import (
	"os"
	"fmt"
	"encoding/json"
	"errors"
)

type Config struct {
	DbConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		Name string `json:"name"`
	} `json:"db_config"`
	HttpConfig struct {
		Listen         string `json:"listen"`
		Port           int    `json:"port"`
		SSLCertPath    string `json:"ssl_cert_path"`
		SSLCertKeyPath string `json:"ssl_cert_key_path"`
	} `json:"http_config"`
	RootCertPath    string `json:"root_cert_path"`
	RootCertKeyPath string `json:"root_cert_private_key_path"`
	CertTTL         int    `json:"cert_ttl"`
}

func GetConfig() *Config {
	config := NewConfig()
	configFilePath, err := FindConfig()
	if err == nil {
		configFile, err := os.Open(configFilePath)
		defer configFile.Close()
		if err != nil {
			Error.Println(fmt.Sprintf("config file found in %s but could not be read: %s", configFilePath, err.Error()))
		}
		jsonParser := json.NewDecoder(configFile)
		jsonParser.Decode(&config)
	} else {
		Info.Println("Config file not found, using default config")
	}

	return config
}

func FindConfig() (string, error) {
	var paths = [] string{
		"config.json",
		"config/application.json",
	}

	var found string
	for _, file := range paths {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			found = file
			break
		}
	}

	if found == "" {
		return "", errors.New("no config files found")
	}

	return found, nil
}

func NewConfig() *Config {
	return &Config{
		DbConfig: struct {
			Host string `json:"host"`
			Port int    `json:"port"`
			Name string `json:"name"`
		}{Host: "localhost", Port: 27017, Name: "crtmgr"},
		HttpConfig: struct {
			Listen         string `json:"listen"`
			Port           int    `json:"port"`
			SSLCertPath    string `json:"ssl_cert_path"`
			SSLCertKeyPath string `json:"ssl_cert_key_path"`
		}{
			Listen:         "",
			Port:           443,
			SSLCertPath:    "server.crt",
			SSLCertKeyPath: "server.key",
		},
		RootCertPath:    "root.crt",
		RootCertKeyPath: "root.key",
		CertTTL:         30,
	}
}
