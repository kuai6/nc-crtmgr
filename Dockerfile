FROM golang:1.8


ADD . /go/src/github.com/kuai6/nc-crtmgr

RUN go get github.com/julienschmidt/httprouter
RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/mgo.v2/bson
RUN go get github.com/mileusna/crontab
RUN go get github.com/sarulabs/di
RUN go get github.com/op/go-logging

RUN cd /go/src/github.com/kuai6/nc-crtmgr && go build && go install

CMD ["/go/bin/nc-crtmgr", "--config=/go/src/github.com/kuai6/nc-crtmgr/docker-config.json"]

EXPOSE 443