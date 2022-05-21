# Envoy ext-proc gRPC filter demo

> Advanced usage of [Envoy ext-proc gRPC filter](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/ext_proc.proto) demonstration.

## Table of contents

- [Concept](#concept)
- [Usage](#usage)
- [Concurrency](#concurrency)

## Concept

This project is testing in depth how you can work with Envoy ext-proc filter, especially when it comes to share data between processing steps.

This provides:
- a [simple golang service](service), just printing received request headers
- a [Envoy proxy](mesh) exposing the service, and enabling an ext-proc filter
- a [golang gRPC external processor](ext-proc) that will be used by Envoy via the ext-proc filter to manipulate the request to / response of the service to add logic 

The intent:
- we want to extract a CSRF value and add it in a dedicated `X-Extracted-Csrf` response header
- this CSRF can come directly form a `X-Csrf` header
- or from a form post, in the `csrf` field

The workflow:
- a request comes arrives on envoy (http://localhost:10000)
- envoy starts a gRPC connection, via the ext-proc filter, to our external processor (:50051)
- envoy streams the request headers to the external processor
- if the request contains a `X-Csrf` header, it's stored in a variable and the request if propagated downstream to the service
- if not, the external processor asks to be streamed with the request body, and the `csrf` post field is stored in a variable, and the request if propagated downstream to the service
- the downstream service processes the request and generates a response
- envoy stream the response to the external processor
- the external processor add a new `X-Extracted-Csrf` response header with the csrf variable content

## Usage

To test the stack, simply start the provided [docker-compose](docker-compose.yaml) stack:
```shell
docker-compose up -d
```

Then you can interact with the service through envoy on [[POST] http://localhost:10000](http://localhost:10000)
- by sending the `X-Csrf` header and check you get it back in the `X-Extracted-Csrf` response header
- or by sending the `csrf` form post field and check you get it back in the `X-Extracted-Csrf` response header


## Concurrency

The main goal of this project is to ENSURE that this pattern does not come with concurrency issues, especially because of the use, external processor side, of high scope variables across gRPC stream handling.

To validate there is actually no issues, the project provides 2 [K6](https://k6.io/) benchmarking scripts:
- [headerBench.js](k6/headerBench.js): to test under traffic the CSRF provided via header
- [bodyBench.js](k6/bodyBench.js): to test under traffic the CSRF provided via form post

To run the scripts, you can find below an example:

```shell
k6 run --vus 50 --iterations 100 k6/headerBench.js
```
