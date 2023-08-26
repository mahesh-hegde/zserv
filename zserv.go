package main

// TODO:
// * Per-file memory limit
// * Unit tests
// * Zip internal paths
// * Buffer entry cache / allocator
// * Config file

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	flag "github.com/spf13/pflag"
)

func checkError(err error, format string, args ...any) {
	if err != nil {
		log.Println(err)
		log.Fatalf(format, args...)
	}
}

type Options struct {
	Port          int
	MaxBufferSize int64
	Verbose       bool
	Root          string
	Host          string
}

var options Options

func verbose(format string, values ...any) {
	if options.Verbose {
		log.Printf(format+"\n", values...)
	}
}

func StartServer(options *Options, zipFS fs.FS) {
	mux := http.NewServeMux()

	hostAddr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	fileserver := http.FileServer(http.FS(zipFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		verbose("\nclient: %v", req.RemoteAddr)
		verbose("request: %v", req.RequestURI)
		fileserver.ServeHTTP(w, req)
	})
	log.Fatal(http.ListenAndServe(hostAddr, mux))
}

func main() {
	var maxBufferSizeStr string
	flag.IntVarP(&options.Port, "port", "p", 8088, "Port to listen on")
	flag.BoolVarP(&options.Verbose, "verbose", "v", false, "Verbose output")
	flag.StringVarP(&options.Root, "root", "r", ".", "Root of the website served relative to ZIP file")
	flag.StringVarP(&options.Host, "host", "h", "127.0.0.1", "Host adress to bind to")
	flag.StringVarP(&maxBufferSizeStr, "max-buffer-size", "Z", "256M", "Maximum file size allowed")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	maxBufferSize, err := parseLimit(maxBufferSizeStr)
	checkError(err, "cannot parse parameter for max buffer size")
	log.Println(maxBufferSize)
	options.MaxBufferSize = maxBufferSize

	path := args[0]
	reader := OpenZipReaderFS(path, options.MaxBufferSize)

	var webfs fs.FS = reader

	if options.Root != "." {
		verbose("opening sub-filesystem at %s", options.Root)
		var err error
		webfs, err = fs.Sub(reader, options.Root)
		checkError(err, "cannot open webserver root!")
	}

	go StartServer(&options, webfs)
	log.Printf("zserv bound to %s:%d\n", options.Host, options.Port)
	fmt.Scanln()
}
