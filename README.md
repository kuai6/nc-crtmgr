# NC Certificate Authority Service

[TOC]

## About

This is private CA (Certificate Authority) Service.

### Usage

First of all we need generate root CA and SSL certificates (you may use your own if you have it)

#### Generate root certificate

```
openssl genrsa -out rootCA.key 2048

openssl req -x509 -new -key rootCA.key -days 365 -out rootCA.crt
```

Put your rootCA.key and rootCA.crt into directory and change ```root_cert_*``` section in config.json

#### Generate SSL

```
openssl genrsa -out server.key 2048

openssl req -new -key server.key -out server.csr

openssl x509 -req -in server.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out server.crt -days 365
```

Put you server.key and server.crt into directory and change ```http_config``` section in config.json

## Configuration

The config file name is config.json

##### Config options:

```db_config``` The MongoDB config section

```root_cert_path``` Path to root certificate

```root_cert_private_key_path``` Path to private key

```http_config``` The HTTP config section, contains host and port to bind and ssl certificate path

```cert_ttl``` Default time to live for generated certificates

```key_rsa_bits``` Generated private kes number bits. Default 2048

```certificate_subject``` Default files to fill subject in generated certificate

##### Config file Example


```
{
  "db_config": {
    "host": "127.0.0.1",
    "port": 27017
  },
  "root_cert_path": "ssl/root/rootCA.crt",
  "root_cert_private_key_path": "ssl/root/rootCA.key",
  "http_config": {
    "listen": "127.0.0.1",
    "port": 8443,
    "ssl_cert_path": "ssl/server.crt",
    "ssl_cert_key_path": "ssl/server.key"
  },
  "cert_ttl": 30,
  "key_rsa_bits": 2048,
  "certificate_subject": {
    "common_name": "nc.ca",
    "country": "RU",
    "province": "Nizhegorodskaya Oblast",
    "locality": "Nizhniy Novgorod",
    "organization": "NC",
    "organizational_unit": "IT Department"
  }
}
```

## Build

### Requirements
- go get github.com/julienschmidt/httprouter
- go get gopkg.in/mgo.v2
- go get gopkg.in/mgo.v2/bson
- go get github.com/mileusna/crontab
- go get github.com/sarulabs/di
- go get github.com/op/go-logging

### Build project
```
$ cd /path/to/project
$ go build
```

## Run

### Command line arguments

```--config=/path/to/config.json``` Path to config file. If not specified the application will try find ```config.json``` into application dir and ```./config/```

```-v``` Verbose flag



## API

#### Request content
Each request contain json structure with required fields ```uid``` and ```did```. Each request must be with header ```Content-type: application/json; charset=UTF-8```. The ```certificate``` fields is optional anf in base64 encode. The ```password``` filed is optional.

#### Response content
Each response contain json structure with required fields ```uid```, ```did```, ```result``` and ```reason```. When error occurs the ```result``` field is set to **false** and the ```reason``` field is set error reason describe. The ```certificate``` and ```private_key``` fields are in base64 encode.


#### Generate certificate request

- Method: POST
- Endpoint: /api/v1/generate
- Post data:
```
{
  "did": "fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "uid": "08cbef46-c6d2-11e7-abc4-cec278b6b50f"
}
```

Response:
```
{
  "uid":"08cbef46-c6d2-11e7-abc4-cec278b6b50f",
  "did":"fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "certificate": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk...",
  "private_key":"LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0...",
  "valid_till":"2017-12-19T12:15:27+03:00",
  "result":true,
  "reason":""
}
```


#### Generate certificate with encrypted private key

- Method: POST
- Endpoint: /api/v1/generate
- Post data:
```
{
  "did": "fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "uid": "08cbef46-c6d2-11e7-abc4-cec278b6b50f",
  "password": "somepass123
}
```

Response:
```
{
  "uid":"08cbef46-c6d2-11e7-abc4-cec278b6b50f",
  "did":"fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "certificate": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk...",
  "private_key":"LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0...",
  "valid_till":"2017-12-19T12:19:27+03:00",
  "result":true,
  "reason":""
}
```


#### Validate certificate

- Method: POST
- Endpoint: /api/v1/validate
- Post data:
```
{
  "did": "fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "uid": "08cbef46-c6d2-11e7-abc4-cec278b6b50f",
  "certificate": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk..."
}
```

Response:
```
{
  "uid":"08cbef46-c6d2-11e7-abc4-cec278b6b50f",
  "did":"fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "result":true,
  "reason":""
}
```

#### Withdrawal certificate

- Method: POST
- Endpoint: /api/v1/validate
- Post data:
```
{
  "uid":"08cbef46-c6d2-11e7-abc4-cec278b6b50f",
  "did":"fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "certificate": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk..."
}
```

Response:

```
{
  "uid":"fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "did":"fc6e1864-c6d1-11e7-abc4-cec278b6b50d",
  "result":true,
  "reason":""
}
```

## Docker image

```
$ sudo docker-compose up -d --build

```
