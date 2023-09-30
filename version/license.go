package version

import (
	"fmt"
	"io"
)

type License struct {
	ModuleName  string
	LicenseName string
	Link        string
}

var Licenses = []License{
	{
		ModuleName:  "httpie-go",
		LicenseName: "MIT License",
		Link:        "https://github.com/HexmosTech/httpie-go/blob/master/LICENSE",
	},
	{
		ModuleName:  "Go",
		LicenseName: "BSD License",
		Link:        "https://golang.org/LICENSE",
	},
	{
		ModuleName:  "aurora",
		LicenseName: "WTFPL",
		Link:        "https://github.com/logrusorgru/aurora/blob/master/LICENSE",
	},
	{
		ModuleName:  "go-isatty",
		LicenseName: "MIT License",
		Link:        "https://github.com/mattn/go-isatty/blob/master/LICENSE",
	},
	{
		ModuleName:  "getopt",
		LicenseName: "BSD License",
		Link:        "https://github.com/pborman/getopt/blob/master/LICENSE",
	},
	{
		ModuleName:  "errors",
		LicenseName: "BSD License",
		Link:        "https://github.com/pkg/errors/blob/master/LICENSE",
	},
	{
		ModuleName:  "bytefmt",
		LicenseName: "Apache License",
		Link:        "https://github.com/cloudfoundry/bytefmt/blob/master/LICENSE",
	},
	{
		ModuleName:  "ewma",
		LicenseName: "MIT License",
		Link:        "https://github.com/VividCortex/ewma/blob/master/LICENSE",
	},
	{
		ModuleName:  "stripansi",
		LicenseName: "MIT License",
		Link:        "https://github.com/acarl005/stripansi/blob/master/LICENSE",
	},
	// {
	// 	ModuleName:  "mpb",
	// 	LicenseName: "Unlicense",
	// 	Link:        "https://github.com/vbauerster/mpb/blob/master/UNLICENSE",
	// },
}

func PrintLicenses(w io.Writer) {
	for _, license := range Licenses {
		fmt.Fprintf(w, "%s:\n  %s\n  %s\n\n",
			license.ModuleName,
			license.LicenseName,
			license.Link,
		)
	}
}
