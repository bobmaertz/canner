BINARY_NAME=canner

build:
	GOARCH=amd64 GOOS=darwin go build -o bin/${BINARY_NAME}-darwin main.go
	GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux main.go
	GOARCH=amd64 GOOS=windows go build -o bin/${BINARY_NAME}-windows main.go

# build:
# 	go build -o bin/canner main.go

run: build
	./${BINARY_NAME}

test_coverage:
 	go test ./... -coverprofile=coverage.out

clean:
	go clean 
	rm bin/*

