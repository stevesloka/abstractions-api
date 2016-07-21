# Makefile for the Docker image stevesloka/abstractions-api
# MAINTAINER: Steve Sloka <steve@stevesloka.com>
# If you update this image please bump the tag value before pushing.

.PHONY: all build container push clean test

TAG = latest
PREFIX = stevesloka

all: container

build: main.go
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -a -installsuffix cgo -o restapi --ldflags '-w' ./main.go

container: build
	docker build -t $(PREFIX)/abstractions-api:$(TAG) .

push:
	docker push $(PREFIX)/abstractions-api:$(TAG)

clean:
	rm -f restapi

test: clean
	godep go test -v --vmodule=*=4
