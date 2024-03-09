package main

// TODO:
// * Auto detect website root
// * Benchmark memory usage when downloading a large file
// * Search endpoint
// --
// * Buffer entry cache / allocator
// * Zip internal paths - if same server is used to serve multiple paths
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
	BufferFiles   bool
	Verbose       bool
	Root          string
	DetectRoot    bool
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
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("%v [%s] %v: %v\n", req.RemoteAddr, req.Method, req.RequestURI, err)
			}
		}()
		fileserver.ServeHTTP(w, req)
	})
	log.Fatal(http.ListenAndServe(hostAddr, mux))
}

func main() {
	var maxBufferSizeStr string
	flag.IntVarP(&options.Port, "port", "p", 8088, "Port to listen on")
	flag.BoolVarP(&options.Verbose, "verbose", "v", false, "Verbose output")
	flag.BoolVarP(&options.BufferFiles, "buffer", "b", false, "Load files completely into memory before serving them")
	flag.StringVarP(&options.Root, "root", "r", ".", "Root of the website served relative to ZIP file")
	flag.BoolVarP(&options.DetectRoot, "detect-root", "R", false,
		"Auto detect root folder as first folder containing multiple entries or an index.html")
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
	options.MaxBufferSize = maxBufferSize

	path := args[0]
	webfs := OpenZipReaderFS(path, &options)

	if options.DetectRoot && options.Root != "." {
		log.Fatalf("Conflicting options: set root and detect root")
	}

	if options.Root != "." {
		verbose("opening sub-filesystem at %s", options.Root)
		var err error
		webfs, err = fs.Sub(webfs, options.Root)
		checkError(err, "cannot open webserver root!")
	}

	if options.DetectRoot {
		sub, err := detectRoot(webfs)
		if err != nil {
			log.Printf("auto-detection of website root failed: %s", err)
		} else {
			webfs = sub
		}
	}

	if options.DetectRoot {
		verbose("Detect root as: %s")
	}

	log.Printf("zserv binding to http://%s:%d\n", options.Host, options.Port)
	StartServer(&options, webfs) 
}
