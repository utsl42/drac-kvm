# drac-kvm ![License badge][license-img]

1. [Overview](#overview)
2. [Description](#description)
3. [Setup](#setup)
4. [Usage](#usage)
5. [Limitations](#limitations)
6. [Development](#development)
7. [Miscellaneous](#miscellaneous)

## Overview

The integrated Dell Remote Access Controller  or DRAC (iDRAC) is an out-of-band
management platform  on certain Dell  servers.  It provides  functionality that
helps you deploy,  update, monitor and maintain Dell PowerEdge  servers with or
without a systems management software agent.

[dell.com](https://www.dell.com/)

A preliminary  implementation of iLO  (Integrated Lights Out) KVM  is available
for version iLO 3 and iLO 4.

[hp.com](https://www.hpe.com/)

## Description

A simple CLI launcher for Dell DRAC and HP iLO KVM sessions


This has been tested on the following Dell servers:

 * 11th Generation (eg: Dell R710 / iDRAC6)

 * 12th Generation (eg: Dell R720 / iDRAC7)

The following HP servers:

 * 7th Generation (eg: HP DL120 G7)

 * 8th Generation (eg: HP DL160 G8)

This has also been tested against a Supermicro based system:

 * X9DR7-LN4F motherboard with ATEN based iKVIM (3.19 fimrware)

## Setup

It requires  that you  have java  installed on  your machine  (specifically the
`javaws` binary).

### Docker

```bash
```

### Go

If you  already have Go  configured on  your system then  you can just  run the
following to quickly install it:

```bash
go get github.com/rockyluke/drac-kvm
```

### Homebrew

If you already  have Homebrew configured on  your system then you  can just run
the following to quickly install it:

```bash
brew tap rockyluke/devops
brew install drac-kvm
```

## Usage

Fed up of logging into the DRAC web interface just to launch a KVM session?
This simple Go program should help ease the pain.

```bash
drac-kvm --help
Usage of drac-kvm
  -h, --host="some.hostname.com": The DRAC host (or IP)
  -j, --javaws="/usr/bin/javaws": The path to javaws binary
  -p, --password=false: Prompt for password (optional, will use 'calvin' if not present)
  -u, --username="": The DRAC username
  -v, --version=-1: iDRAC version (6 or 7) (supermicro ikvm as 1)
```

### Example using default dell credentials (root/calvin)

```bash
drac-kvm -h 10.25.1.100
2014/06/26 16:01:11 Detecting iDRAC version...
2014/06/26 16:01:11 Found iDRAC version 7
2014/06/26 16:01:11 Launching DRAC KVM session to 10.25.1.100
```

### Example using custom credentials

```bash
drac-kvm -h 10.25.1.100 -u bob -p
Password: **********
2014/06/26 16:01:11 Detecting iRAC version...
2014/06/26 16:01:11 Found iDRAC version 7
2014/06/26 16:01:11 Launching DRAC KVM session to 10.25.1.100
```

### Configuration file

You can create a configuration file

```bash
cat ~/.drackvmrc
# Override the hardcoded defaults for username and password.
# Useful if your environment has consistent usernames and
# passwords for the KVMs.
[defaults]
username = foo
password = bar

[192.168.0.42]
username = foo
password = bar

[web-1]
host = 10.33.0.1
username = root
password = password4root

[web-2]
host = 10.33.0.2
username = root
```

```bash
# Use IP, username and password from drackvmrc for this ip
drac-kvm -h 192.168.0.42
2014/06/26 16:01:11 Detecting iRAC version...
2014/06/26 16:01:11 Found iDRAC version 7
2014/06/26 16:01:11 Launching DRAC KVM session to 192.168.0.42

# Same for host alias web-1
drac-kvm -h web-1
2014/06/26 16:01:11 Detecting iRAC version...
2014/06/26 16:01:11 Found iDRAC version 7
2014/06/26 16:01:11 Launching DRAC KVM session to 10.33.0.1

# Specify -p option in order to do not use defaut password
drac-kvm -h web-2 -p
Password: **********
2014/06/26 16:01:11 Detecting iRAC version...
2014/06/26 16:01:11 Found iDRAC version 7
2014/06/26 16:01:11 Launching DRAC KVM session to 10.33.0.2
```

## Credits

@jamesdotcuff's helpful blog post:

http://blog.jcuff.net/2013/10/fun-with-idrac.html

## Development

Feel free to contribute on GitHub.

```
    ╚⊙ ⊙╝
  ╚═(███)═╝
 ╚═(███)═╝
╚═(███)═╝
 ╚═(███)═╝
  ╚═(███)═╝
   ╚═(███)═╝
```

[license-img]: https://img.shields.io/badge/license-Apache%202.0-blue.svg
