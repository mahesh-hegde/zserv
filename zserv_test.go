package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func downloadFile(port int, path string) []byte {
	link := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", port),
		Path:   path,
	}
	response, err := http.Get(link.String())
	checkError(err, "cannot get file from HTTP")
	bytes, err := io.ReadAll(response.Body)
	checkError(err, "cannot copy file from HTTP")
	return bytes
}

func TestBasicDownload(t *testing.T) {
	port := 9688
	repeated100 := strings.Repeat("<p>ABCDEFGHIJKLMNOPQRSTUVWXYZ</p>", 100)
	repeated10000 := strings.Repeat("<p>ABCDEFGHIJKLMNOPQRSTUVWXYZ</p>", 10000)
	entries := []entry{
		{"small.html", fmt.Sprintf("<html><body>%s</body></html>\n", repeated100)},
		{"big.html", fmt.Sprintf("<html><body>%s</body></html>\n", repeated10000)},
		{"nest/ed.html", "<html><body>Hello</body></html>"},
		{"nest/ed/nest/ed.html", "alwiehgweioghlafknewi;foghiEGNLAERI"},
	}
	for index, fs := range createBothTestZipReaderFS(entries) {
		t.Run(reflect.TypeOf(fs).Name(), func(t *testing.T) {
			go StartServer(&Options{Port: port + index, Host: "127.0.0.1",
				Root: "."}, fs)
			for _, entry := range entries {
				bytes := downloadFile(port, entry.Name)
				assert.Equal(t, entry.Body, string(bytes), "Download %s",
					entry.Name)
			}
		})
	}
}
