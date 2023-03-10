# canner
Canner is a server that provides configurable mock responses for http endpoints


## Building 
```shell
make build 
```

## Running 

1. Create a config.yml in the conf directory 
```yaml
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
2. Run
```shell 
make run  
```
