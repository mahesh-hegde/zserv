package main

import (
	"os"
	"fmt"
	"archive/zip"
	"net/http"
)

func checkError(err error, message string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, message)
		os.Exit(2);
	}
}

func main() {
	var _i string
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s file\n", os.Args[0])
		os.Exit(1);
	}
	path := os.Args[1]
	z, err := zip.OpenReader(path)
	checkError(err, "cannot open zip file.")
	http.Handle("/", http.FileServer(http.FS(z)))
	go http.ListenAndServe("127.0.0.1:8080", nil)
	fmt.Println("ZServ is running on port 8080, press enter to stop")
	fmt.Scanln(&_i)
}

