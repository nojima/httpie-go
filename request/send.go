package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nojima/httpie-go/input"
	"github.com/pkg/errors"
)

func SendRequest(request *input.Request) error {
	client, err := buildHttpClient()
	if err != nil {
		return err
	}
	r, err := buildHttpRequest(request)
	if err != nil {
		return err
	}

	resp, err := client.Do(r)
	if err != nil {
		return errors.Wrap(err, "sending HTTP request")
	}
	defer resp.Body.Close()

	printResponseHeader(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}
	fmt.Printf("%s", body)

	return nil
}

func buildHttpClient() (*http.Client, error) {
	client := http.Client{
		// Do not follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}
	return &client, nil
}

func printResponseHeader(resp *http.Response) {
	fmt.Printf("%s %s\n", resp.Proto, resp.Status)
	for name, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}
	fmt.Println("")
}
