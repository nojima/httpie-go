// +build windows

package flags

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

func askPassword() (string, error) {
	fmt.Fprintf(os.Stderr, "Password: ")
	fd := int(os.Stdin.Fd())
	password, err := terminal.ReadPassword(fd)
	if err != nil {
		return "", errors.Wrap(err, "failed to read password from terminal")
	}
	fmt.Fprintln(os.Stderr)
	return string(password), nil
}
