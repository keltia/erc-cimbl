
# Main Makefile for labs2pg
#
# Copyright 2015 © by Ollivier Robert for the EEC
#

GOBIN=   ${GOPATH}/bin

SRCS= config.go mail.go main.go parse.go path.go url.go
SRCSW= config_windows.go
SRCSU= config_unix.go

OPTS=	-ldflags="-s -w" -v

all: erc-cimbl erc-cimbl.exe

erc-cimbl: ${SRCS} ${SRCSU}
	go build ${OPTS}

erc-cimbl.exe: ${SRCS} ${SRCSW}
	GOOS=windows go build ${OPTS}

test:
	go test -v

install: erc-cimbl
	go install ${OPTS}

clean:
	go clean -v

push:
	git push --all
	git push --tags
	git push --all backup
	git push --tags backup
	git push --all upstream
	git push --tags upstream
