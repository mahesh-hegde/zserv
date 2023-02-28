package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
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
		fmt.Fprintf(os.Stderr, format+"\n", values...)
	}
}

type ZipReader struct {
	*zip.Reader
}

type BufferedZipEntry struct {
	fs.File
	*bytes.Reader
}

func (b *BufferedZipEntry) Read(p []byte) (n int, err error) {
	return b.Reader.Read(p)
}

func (b *BufferedZipEntry) ReadDir(n int) ([]fs.DirEntry, error) {
	if rdf, ok := b.File.(fs.ReadDirFile); ok {
		return rdf.ReadDir(n)
	}
	return nil, fmt.Errorf("fs.File instance does not implement ReadDir")
}

func NewBufferedZipEntry(f fs.File) *BufferedZipEntry {
	if f == nil {
		return nil
	}
	var buf bytes.Buffer
	io.Copy(&buf, f)
	return &BufferedZipEntry{f, bytes.NewReader(buf.Bytes())}
}

var _ io.ReadSeeker = &BufferedZipEntry{}
var _ fs.File = &BufferedZipEntry{}

func (z ZipReader) Open(name string) (fs.File, error) {
	verbose("open: %s", name)
	f, err := z.Reader.Open(name)
	if err != nil {
		verbose("error opening %s: %v", name, err)
		return nil, err
	}
	return NewBufferedZipEntry(f), nil
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
	f, _ := os.Open(path)
	fstat, _ := f.Stat()
	z_, err := zip.NewReader(f, fstat.Size())
	checkError(err, "cannot open zip file.")
	z := ZipReader{z_}

	var webfs fs.FS = z

	if options.Root != "." {
		verbose("opening sub-filesystem at %s", options.Root)
		var err error
		webfs, err = fs.Sub(z, options.Root)
		checkError(err, "cannot open webserver root!")
	}

	hostAddr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	fileserver := http.FileServer(http.FS(webfs))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		verbose("\nclient: %v", req.RemoteAddr)
		verbose("request: %v", req.RequestURI)
		fileserver.ServeHTTP(w, req)
	})
	verbose("starting server")
	verbose("host: %s, port %d", options.Host, options.Port)
	fmt.Printf("ZServ running on port %d\n", options.Port)
	go func() {
		log.Fatal(http.ListenAndServe(hostAddr, nil))
	}()
	fmt.Scanln()
}
