package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

func checkError(err error, message string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, message)
		os.Exit(2)
	}
}

type Options struct {
	Port    int
	Verbose bool
	Root    string
	Host    string
}

var options Options

func verbose(format string, values ...any) {
	if options.Verbose {
		fmt.Fprintf(os.Stderr, format + "\n", values...)
	}
}

func main() {
	flag.IntVar(&options.Port, "port", 8080, "Port to listen on")
	flag.BoolVar(&options.Verbose, "verbose", false, "Verbose output")
	flag.StringVar(&options.Root, "root", ".", "Root of the website served relative to ZIP file")
	flag.StringVar(&options.Host, "host", "127.0.0.1", "Host adress to bind to")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	path := args[0]
	z, err := zip.OpenReader(path)
	checkError(err, "cannot open zip file.")

	var webfs fs.FS = z

	if options.Root != "." {
		verbose("opening sub-filesystem at %s", options.Root)
		var err error
		webfs, err = fs.Sub(z, options.Root)
		checkError(err, "cannot open webserver root!")
	}

	hostAddr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	fileserver := http.FileServer(http.FS(webfs))
	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request) {
		verbose("client: %v", req.RemoteAddr)
		verbose("request: %v\n", req.RequestURI)
		fileserver.ServeHTTP(w, req)
	})
	verbose("starting server")
	verbose("host: %s, port %d", options.Host, options.Port)
	fmt.Printf("ZServ running on port %d\n", options.Port)
	go log.Fatal(http.ListenAndServe(hostAddr, nil))
	fmt.Scanln()
}

