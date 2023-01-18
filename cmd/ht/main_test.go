package main

import (
	"os"
	"testing"
)

func TestGooglePing(t *testing.T) {
	os.Args = []string{"./ht", "http://localhost:8080/get"}
	main()
}
