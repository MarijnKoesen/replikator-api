# Replikator API

[![Go Report Card](https://goreportcard.com/badge/github.com/marijnkoesen/replikator-api)](https://goreportcard.com/report/github.com/marijnkoesen/replikator-api)
[![Build Status](https://travis-ci.org/MarijnKoesen/replikator-api.svg?branch=master)](https://travis-ci.org/MarijnKoesen/replikator-api)

This project adds a REST API around the Replikator that makes it possible to manage the replikated databases using a
REST api.

```
Usage of replikator-api:
	Restfull Replikator API server
Options:
  -l, --listen=:8080                      listen address
  -r, --replikator="sudo replikator-ctl"  Path to replikator-ctl
  -s, --secret=                           CORS secret, minimal length 20 chars
  -h, --help                              Show usage message
  --version                               Show version
```

## Run locally

Usage:

```bash
$ go get github.com/MarijnKoesen/replikator-api
$ go build github.com/MarijnKoesen/replikator-api
$ ./replikator-api

# when you don't have replikator-ctl available
$ ./replikator-api -r echo
```

From source:

```bash
$ git clone git@github.com:MarijnKoesen/replikator-api.git
$ go run . -r echo
```

```bash
> curl -XPUT localhost:8080/replikator/foo -vvv
*   Trying 127.0.0.1:8080...
* Connected to localhost (127.0.0.1) port 8080 (#0)
> PUT /replikator/foo HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.79.1
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Access-Control-Allow-Origin: *
< Content-Type: application/json
< Date: Wed, 26 Oct 2022 11:22:50 GMT
< Content-Length: 27
<
--output json --create foo
* Connection #0 to host localhost left intact
```

## Installation on server

Steps to install replikator-api on an ubuntu server:

1) Install replikator-ctl in `/usr/bin/replikator-ctl`

2) Create a `replikator` user/group


```bash
# groupadd replikator
# useradd -r -s /bin/nologin replikator2
```

3) Create file `/etc/sudoers.d/replikator` with content:

```sudoers
replikator ALL=(ALL) NOPASSWD:/usr/bin/replikator-ctl
```

4) Install replikator-api binary

```bash
# go get github.com/MarijnKoesen/replikator-api
# go build -o /usr/bin/replikator-api  github.com/MarijnKoesen/replikator-api
```

5) Create a replikator service

Create file `/etc/systemd/system/replikator-api.service`

```service
[Unit]
Description=Replikator API
ConditionPathExists=/usr/bin/replikator-api
After=network.target

[Service]
Type=simple
User=replikator
Group=replikator
LimitNOFILE=1024

KillMode=process

Restart=on-failure
RestartSec=10

ExecStart=/usr/bin/replikator-api

[Install]
WantedBy=multi-user.target
```

Now install and enable the service:

```bash
# systemctl daemon-reload
# systemctl enable replikator-api
# systemctl start replikator-api
```
