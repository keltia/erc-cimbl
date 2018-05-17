# Main Makefile for labs2pg
#
# Copyright 2015 © by Ollivier Robert for the EEC
#

GOBIN=   ${GOPATH}/bin

SRCS= config.go mail.go main.go parse.go path.go url.go utils.go
SRCSW= config_windows.go
SRCSU= config_unix.go

OPTS=	-ldflags="-s -w" -v

PROG=	erc-cimbl
BIN=	${PROG}
EXE=	${PROG}.exe

all: ${BIN}

${BIN}: ${SRCS} ${SRCSU}
	go build ${OPTS}

${EXE}: ${SRCS} ${SRCSW}
	GOOS=windows go build ${OPTS}

test:
	go test -v

install: ${BIN}
	go install ${OPTS}

lint:
	gometalinter .

clean:
	go clean -v

push:
	git push --all
	git push --tags
	git push --all backup
	git push --tags backup
