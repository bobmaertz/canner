# canner
Canner is a server that provides configurable mock responses


## Building 
```shell
go build 
```

## Running 

- Create a config.yml in the conf directory 
```sh
server:
  port: 8450

matchers:
  - request:
      path: /hello/world
      headers:
        ExampleHeaderToMatch: "Example"
    response:
     body: 'hello!'
     statusCode: 200
     headers:
       Content-Type: "text/plain; charset=utf-8"
       
```
