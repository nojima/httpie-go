package output

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type FileWriter struct {
	fullPath string
}

func NewFileWriter(url *url.URL, options *Options) *FileWriter {
	var fullPath string

	if options.OutputFile == "" {
		fullPath = fmt.Sprintf("./%s", filepath.Base(url.Path))
	} else {
		fullPath = options.OutputFile
	}

	if !options.Overwrite {
		fullPath = makeNonOverlappingFilename(fullPath)
	}

	return &FileWriter{
		fullPath: fullPath,
	}
}

func makeNonOverlappingFilename(path string) string {
	_, err := os.Stat(path)
	if err == nil {
		re := regexp.MustCompile(`\.(\d+)$`)
		newPath := re.ReplaceAllStringFunc(path, func(index string) string {
			i, err := strconv.Atoi(strings.TrimPrefix(index, "."))
			if err != nil {
				panic(err)
			}
			i++
			return fmt.Sprintf(".%d", i)
		})
		if path == newPath {
			path = fmt.Sprintf("%s.%d", path, 1)
		} else {
			path = newPath
		}
		path = makeNonOverlappingFilename(path)
	}
	return path
}

// func (f *FileWriter) Download(resp *http.Response) error {
// 	// Create new progress bar
// 	pb := mpb.New(mpb.WithWidth(60))

// 	// Create file
// 	file, err := os.Create(f.fullPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	// Parameters of th new progress bar
// 	bar := pb.AddBar(resp.ContentLength,
// 		mpb.PrependDecorators(
// 			decor.CountersKiloByte("% .2f / % .2f "),
// 			decor.AverageSpeed(decor.UnitKB, "(% .2f)"),
// 		),
// 		mpb.AppendDecorators(
// 			decor.Percentage(),
// 			decor.Name(" - "),
// 			decor.Elapsed(decor.ET_STYLE_GO, decor.WC{W: 4}),
// 			decor.Name(" - "),
// 			decor.OnComplete(
// 				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "done",
// 			),
// 		),
// 	)

// 	// Update progress bar while writing file
// 	_, err = io.Copy(file, bar.ProxyReader(resp.Body))
// 	if err != nil {
// 		return err
// 	}

// 	pb.Wait()

// 	return nil
// }

func (f *FileWriter) Download(resp *http.Response) error {
	// Create file
	file, err := os.Create(f.fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get content length for progress calculation
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		// If content length is not provided, just copy the content without progress
		_, err = io.Copy(file, resp.Body)
		return err
	}

	// Buffer for reading chunks of data
	buf := make([]byte, 4096)
	var totalRead int64

	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// Write the read chunk to the file
		_, err = file.Write(buf[:n])
		if err != nil {
			return err
		}

		// Update total read and print progress
		totalRead += int64(n)
		percentage := (totalRead * 100) / contentLength
		fmt.Printf("\rProgress: %d%%", percentage)
	}

	fmt.Println("\nDownload complete!")
	return nil
}

func (f *FileWriter) Filename() string {
	return filepath.Base(f.fullPath)
}
