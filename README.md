# Go load balancer

## Overview

An application load balancer built in Go.

## Usage

### Build

1. Build the back-end service:
```bash
$ go build -o be cmd/be/main.go
```

2. Build the load balancer service:
```bash
$ go build -o be cmd/lb/main.go
```

### Run

1. Run your back-end services locally:
```bash
./be [port]
```

Example
```bash
$ ./be 5001
```
```bash
$ ./be 5002
```


2. Run your load balancer service locally:
```bash
$ ./lb [service1-address] ... [serviceN-address]
```

Example

```bash
$ ./lb 127.0.0.1:5001 127.0.0.1:5002
```

3. Send a request to the load balancer

Example
```bash
$ curl -X GET http://127.0.0.1:5000/
hello 127.0.0.1:5001
```
While on the back-end service to which the request gets forwarded to:

```bash
Received request from 
GET / HTTP/1.1
Host 127.0.0.1:5001:
User-Agent: Go-http-client/1.1
Accept: gzip
```

## Acknowledgement

Coding Challenge ["Build Your Own Load Balancer"](https://codingchallenges.fyi/challenges/challenge-load-balancer). Go check out John Crickett's [Coding Challenges newsletter](https://codingchallenges.fyi/) for more inspiring challenges.
