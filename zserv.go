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

func checkError(err error, format string, args ...any) {
	if err != nil {
		log.Println(err)
		log.Fatalf(format, args...)
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
		log.Printf(format+"\n", values...)
	}
}

type ZipReaderFS struct {
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

var _ fs.FS = &ZipReaderFS{}

func (z ZipReaderFS) Open(name string) (fs.File, error) {
	verbose("open: %s", name)
	f, err := z.Reader.Open(name)
	if err != nil {
		verbose("error opening %s: %v", name, err)
		return nil, err
	}
	return NewBufferedZipEntry(f), nil
}

func OpenZipReaderFS(path string) fs.FS {
	f, err := os.Open(path)
	checkError(err, "cannot open input file")
	fstat, err := f.Stat()
	checkError(err, "Cannot stat input file")
	return GetZipReaderFS(f, fstat.Size())
}

func GetZipReaderFS(reader io.ReaderAt, size int64) ZipReaderFS {
	zipReader, err := zip.NewReader(reader, size)
	checkError(err, "cannot open input ZIP file")
	return ZipReaderFS{zipReader}
}

// StartServer spawns the server in a separate goroutine and returns.
func StartServer(options *Options, zipFS fs.FS) {
	mux := http.NewServeMux()

	hostAddr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	fileserver := http.FileServer(http.FS(zipFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		verbose("\nclient: %v", req.RemoteAddr)
		verbose("request: %v", req.RequestURI)
		fileserver.ServeHTTP(w, req)
	})
	go func() {
		log.Fatal(http.ListenAndServe(hostAddr, mux))
	}()

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
	reader := OpenZipReaderFS(path)

	var webfs fs.FS = reader

	if options.Root != "." {
		verbose("opening sub-filesystem at %s", options.Root)
		var err error
		webfs, err = fs.Sub(reader, options.Root)
		checkError(err, "cannot open webserver root!")
	}
	StartServer(&options, webfs)
	log.Printf("zserv bound to %s:%d\n", options.Host, options.Port)
	fmt.Scanln()
}
