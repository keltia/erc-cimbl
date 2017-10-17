erc-cimbl
============

[![Build Status](https://travis-ci.org/keltia/erc-cimbl.svg?branch=master)](https://travis-ci.org/keltia/erc-cimbl)
[![GoDoc](http://godoc.org/github.com/keltia/erc-cimbl?status.svg)](http://godoc.org/github.com/keltia/erc-cimbl)

This is a small utility that parse files sent by the CERT-EU.  These weekly files contains a few URLs and filenames everyone want to block at the firewall/web content proxies.

This will identify what we want and build a mail to be cut&pasted (and later sent directly).

## Requirements

* Go >= 1.8

## Usage

SYNOPSIS
```
erc-cimbl [-M] [-P] [-U] [-v] CIMBL-nnn-CERTS.csv

NOTE: the file must have this format.
```

OPTIONS

| Option  | Default | Description|
| ------- |---------|------------|
| -M      | false   | Actually send mail | ** NOT IMPLEMENTED **
| -P      | false   | Do not check filenames |
| -U      | false   | Do not check URLs |
| -v      | false   | Be verbose |

## Using behind a web Proxy

Linux/Unix:
```
    export HTTP_PROXY=[http://]host[:port] (sh/bash/zsh)
    setenv HTTP_PROXY [http://]host[:port] (csh/tcsh)
```

Do not forget to run these commands to create the configuration file:

    mkdir -p $HOME/.config/erc-cimbl

and to put the *edited* content of the default `config.toml` there.

Windows — ONLY if you use a version < 0.4 as I can not use `cgo` when cross-building.
```
    set HTTP_PROXY=[http://]host[:port]
```

The rules of Go's `ProxyFromEnvironment` apply (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`, lowercase variants allowed).

Do not forget to run these commands to create the configuration file:

    mkdir %APPLOCALDATA%\DG-CSS\CIMBL

and to put the *edited* content of the default `config.toml` there.

## BUGS

v0.4 started supporting direct GPGME decryption and this does not work on Windows.

## License

The [BSD 2-Clause license][bsd].

# Feedback

We welcome pull requests, bug fixes and issue reports.

Before proposing a large change, first please discuss your change by raising an issue.
