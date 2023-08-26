package main

// TODO:
// * Per-file memory limit
// * Unit tests
// * Zip internal paths
// * Buffer entry cache / allocator
// * Config file

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
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

type ZipReaderFS struct {
	*zip.Reader
	maxBufferSize int64
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

var errBufferSizeExceeded = fmt.Errorf("size exceeds maximum allowed buffer size")

func (z ZipReaderFS) Open(name string) (fs.File, error) {
	verbose("open: %s", name)
	f, err := z.Reader.Open(name)
	if err != nil || f == nil {
		verbose("error opening %s: %v", name, err)
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() && stat.Size() >= z.maxBufferSize {
		verbose("Buffer size exceeded for entry: %s, size: %d", name, stat.Size())
		return nil, errBufferSizeExceeded
	}
	return NewBufferedZipEntry(f), nil
}

func OpenZipReaderFS(path string, maxBufferSize int64) fs.FS {
	f, err := os.Open(path)
	checkError(err, "cannot open input file")
	fstat, err := f.Stat()
	checkError(err, "Cannot stat input file")
	return GetZipReaderFS(f, fstat.Size(), maxBufferSize)
}

func GetZipReaderFS(reader io.ReaderAt, size int64,
	maxBufferSize int64) ZipReaderFS {
	zipReader, err := zip.NewReader(reader, size)
	checkError(err, "cannot open input ZIP file")
	return ZipReaderFS{zipReader, maxBufferSize}
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
