![Build](https://github.com/bobmaertz/canner/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobmaertz/canner)](https://goreportcard.com/report/github.com/bobmaertz/canner)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/bobmaertz/canner/blob/master/LICENSE.md)


Canner is a server that provides configurable mock http server written in go. This server supports 


# Features
 - Coming Soon: Support for simulation of network errors
 - Support for injection of a latency delay into the response. Currently simple and random latency modes are supported. Simple is a static duration delay while random will choose a latency up to the value provided in the configuration. 
 - Supports matching body, method, and headers to serve up a response 

# Getting Started

## Installation 

The easiest way to get started is to to `go get github.com/bobmaertz/canner` and create a configuration file.


## Building 
```shell
make build 
```

## Running 

Before running, create a conf/ directory next to where canner will run and copy your config.yml there. 

Example Config file: 
```yaml
server:
  port: 8450

matchers:
  - request:
      method: Get 
      path: /hello/world
      headers:
        ExampleHeaderToMatch: "Example"
    response:
     body: 'hello!'
     statusCode: 200
     headers:
       Content-Type: "text/plain; charset=utf-8"
```

Running instructions 

```shell 
make run  
 
OR 

./bin/canner-{os} 
```


## TODO / Coming Soon: 
- Support for config file location other than co-located conf/config.yml 
- Support for sample config generation based on a URL
- Support for request/response capture 
