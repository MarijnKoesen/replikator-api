# Replikator API

[![Go Report Card](https://goreportcard.com/badge/github.com/marijnkoesen/replikator-api)](https://goreportcard.com/report/github.com/marijnkoesen/replikator-api)
[![Build Status](https://travis-ci.org/MarijnKoesen/replikator-api.svg?branch=master)](https://travis-ci.org/MarijnKoesen/replikator-api)

This project adds a REST API around the Replikator that makes it possible to manage the replikated databases using a
REST api.


## Run locally

Usage:

```bash
$ go get github.com/MarijnKoesen/replikator-api
$ go build github.com/MarijnKoesen/replikator-api
$ ./replikator-api
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
