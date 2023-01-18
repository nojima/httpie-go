package main

import (
	"os"
	"testing"
)

func TestGooglePing(t *testing.T) {
	os.Args = []string{"./ht", "https://google.com"}
	main()
}
