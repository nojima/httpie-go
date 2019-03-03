EXE_NAME=hg

build:
	go build -o $(EXE_NAME) ./cmd/hg

fmt:
	go fmt ./...

test:
	go test ./...

clean:
	rm -vf ./$(EXE_NAME)

.PHONY: build test clean
