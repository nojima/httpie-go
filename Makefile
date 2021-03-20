EXE_NAME=ht

build:
	go build -o $(EXE_NAME) ./cmd/ht

build-termux:
	env CGO_ENABLED=0 go build -o $(EXE_NAME) -tags='androiddnsfix' ./cmd/ht

install:
	go install ./cmd/ht

fmt:
	go fmt ./...

test:
	go test ./...

clean:
	rm -vf ./$(EXE_NAME)

.PHONY: build test clean
