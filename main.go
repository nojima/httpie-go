package httpie

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/nojima/httpie-go/exchange"
	"github.com/nojima/httpie-go/flags"
	"github.com/nojima/httpie-go/input"
	"github.com/nojima/httpie-go/output"
	"github.com/pkg/errors"
)

func Main() error {
	// Parse flags
	args, usage, optionSet, err := flags.Parse(os.Args)
	if err != nil {
		return err
	}
	inputOptions := optionSet.InputOptions
	exchangeOptions := optionSet.ExchangeOptions
	outputOptions := optionSet.OutputOptions

	// Parse positional arguments
	in, err := input.ParseArgs(args, os.Stdin, &inputOptions)
	if _, ok := errors.Cause(err).(*input.UsageError); ok {
		usage.PrintUsage(os.Stderr)
		return err
	}
	if err != nil {
		return err
	}

	// Send request and receive response
	if err := Exchange(in, &exchangeOptions, &outputOptions); err != nil {
		return err
	}

	return nil
}

func Exchange(in *input.Input, exchangeOptions *exchange.Options, outputOptions *output.Options) error {
	// Prepare printer
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	printer := output.NewPrinter(writer, outputOptions)

	// Build HTTP request
	request, err := exchange.BuildHTTPRequest(in, exchangeOptions)
	if err != nil {
		return err
	}

	// Print HTTP request
	if outputOptions.PrintRequestHeader || outputOptions.PrintRequestBody {
		// `request` does not contain HTTP headers that HttpClient.Do adds.
		// We can get these headers by DumpRequestOut and ReadRequest.
		dump, err := httputil.DumpRequestOut(request, true)
		if err != nil {
			return err // should not happen
		}
		r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(dump)))
		if err != nil {
			return err // should not happen
		}
		defer r.Body.Close()

		// ReadRequest deletes Host header. We must restore it.
		if request.Host != "" {
			r.Header.Set("Host", request.Host)
		} else {
			r.Header.Set("Host", request.URL.Host)
		}

		if outputOptions.PrintRequestHeader {
			if err := printer.PrintRequestLine(r); err != nil {
				return err
			}
			if err := printer.PrintHeader(r.Header); err != nil {
				return err
			}
		}
		if outputOptions.PrintRequestBody {
			if err := printer.PrintBody(r.Body, r.Header.Get("Content-Type")); err != nil {
				return err
			}
		}
		fmt.Fprintln(writer)
		writer.Flush()
	}

	// Send HTTP request and receive HTTP request
	httpClient, err := exchange.BuildHTTPClient(exchangeOptions)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "sending HTTP request")
	}
	defer resp.Body.Close()

	if outputOptions.Download {
		file := output.NewFileWriter(in.URL, outputOptions)

		if outputOptions.PrintResponseHeader {
			if err := printer.PrintStatusLine(resp.Proto, resp.Status, resp.StatusCode); err != nil {
				return err
			}
			if err := printer.PrintHeader(resp.Header); err != nil {
				return err
			}
			writer.Flush()
		}

		if err := printer.PrintDownload(resp.ContentLength, file.Filename()); err != nil {
			return err
		}
		writer.Flush()

		if err = file.Download(resp); err != nil {
			return err
		}

	} else {
		if outputOptions.PrintResponseHeader {
			if err := printer.PrintStatusLine(resp.Proto, resp.Status, resp.StatusCode); err != nil {
				return err
			}
			if err := printer.PrintHeader(resp.Header); err != nil {
				return err
			}
			writer.Flush()
		}
		if outputOptions.PrintResponseBody {
			if err := printer.PrintBody(resp.Body, resp.Header.Get("Content-Type")); err != nil {
				return err
			}
		}
	}

	return nil
}
