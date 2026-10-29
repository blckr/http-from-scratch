# CAP: HTTP from Scratch

This is the base for the training "HTTP from Scratch".
Purpose of this training is to deepen Go development and understanding of the HTTP protocol.

## Intro

The HyperText Transfer Protocol (HTTP) is the basis for the „Internet“.

In this training, we'll implement the still widely supported and used version 1.1, e.g.

```http
GET /index.html HTTP/1.1
Host: www.example.com
User-Agent: Browser
```

returns

```
HTTP/1.1 200 OK
Server: nginx/1.15.8
Date: Fri, 09 Apr 2021 13:34:03 GMT
Content-Type: text/html
Content-Length: 235
<html>
  <head>
  <title>Hello World</title>
</head>
<body>
  Hello World
</body>
</html>
```

## The Task (Day 1)

* Read [the RFC](https://www.rfc-editor.org/info/rfc9112/)
* Implement a basic HTTP server that
  * can handle client connections
  * parse requests
  * write responses

Until this point, it is fine to use the http.DefaultServeMux.


# Deployment Guide

## Prerequisites

* Docker Desktop with Kubernetes enabled (`kubectl config use-context docker-desktop`)
* Nix with active flakes support

## Initial Setup

```bash
# 1. Enter the reproducible environment
nix develop

# 2. Deploy Ingress, VictoriaMetrics, and the Go server
make setup

# 3. Verify connection
curl http://localhost/test
```

## Development & Re-deployment

Run this after making changes to the Go source code or Kubernetes manifests:

```bash
make deploy-local
```

*Builds the Docker image, applies `k8s/` resources, and triggers a rolling restart of the pod.*

## Accessing Monitoring

```bash
# Grafana (Credentials: admin / admin)
kubectl port-forward -n monitoring svc/vmstack-grafana 3000:80
# -> Open: http://localhost:3000

# VictoriaMetrics UI (VMUI)
kubectl port-forward -n monitoring svc/vmsingle-vmstack-victoria-metrics-k8s-stack 8428:8428
# -> Open: http://localhost:8428/vmui
```
