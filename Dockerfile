FROM golang:1.14 as builder

LABEL maintainer="kyawmyintthein"

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -o bin/application main.go

FROM ubuntu:20.04

RUN apt-get update \
  && apt-get install -y ca-certificates curl\
  && update-ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && mkdir -p /go/app

WORKDIR /go/app

COPY --from=builder /go/src/app/bin/application /go/app/
COPY configuration.json .
EXPOSE 8000 8090

ENTRYPOINT ["/go/app/application"]

CMD [ "run", "-c", "/etc/krakend/krakend.json" ]


