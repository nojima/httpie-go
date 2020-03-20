EXE_NAME=ht

build:
	go build -o $(EXE_NAME) ./cmd/ht

install:
	go install ./cmd/ht

fmt:
	go fmt ./...

test:
	go test ./...

clean:
	rm -vf ./$(EXE_NAME)

.PHONY: build test clean
