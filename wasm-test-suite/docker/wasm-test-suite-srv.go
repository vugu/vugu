package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	httpListen := flag.String("http-listen", "127.0.0.1:8846", "HTTP host:port to listen on")
	flag.Parse()

	dirName, err := ioutil.TempDir("", "wasm-test-suite")
	if err != nil {
		panic(err)
	}

	var tsrv TSrv
	tsrv.BaseDir = dirName

	s := &http.Server{
		Addr:           *httpListen,
		Handler:        &tsrv,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("HTTP server listening at %q", *httpListen)
	log.Fatal(s.ListenAndServe())

}

// TSrv is our test server.
type TSrv struct {
	BaseDir string
}

// ServeHTTP implements http.Handler
func (s *TSrv) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	cleanPath := path.Clean("/" + r.URL.Path)

	switch {

	// accept tar.gz file upload and unpack to temp dir
	case cleanPath == "/upload" && r.Method == "POST":
		err := r.ParseMultipartForm(50 * 1024 * 1024) // accept up to 50MB
		if err != nil {
			panic(err)
		}

		file, header, err := r.FormFile("archive")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		log.Printf("Got upload of %q (size=%d)", header.Filename, header.Size)

		dir, err := ioutil.TempDir(s.BaseDir, "fs")
		if err != nil {
			panic(err)
		}

		gr, err := gzip.NewReader(file)
		if err != nil {
			panic(err)
		}
		tr := tar.NewReader(gr)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break // end of archive
			}
			if err != nil {
				panic(err)
			}

			// just skip directories to keep it simple
			if hdr.FileInfo().IsDir() {
				continue
			}

			// close each file as we go
			func() {

				// calc path and make sure parent dir exists
				outPath := filepath.Join(dir, path.Clean("/"+hdr.Name))
				err := os.MkdirAll(filepath.Dir(outPath), 0755)
				if err != nil {
					panic(err)
				}

				// create output file
				outFile, err := os.Create(outPath)
				if err != nil {
					panic(err)
				}
				defer outFile.Close()

				// copy from tar to target file
				_, err = io.Copy(outFile, tr)
				if err != nil {
					panic(err)
				}

			}()

		}

		w.Header().Set("Content-Type", "application/json")
		dirPart := path.Base(path.Clean(dir))
		fmt.Fprintf(w, `{"path":"/%s/","id":"%s"}`, dirPart, dirPart)
		return

	// For paths that don't look like directories, which do not have a file extension,
	// we send them to /index.html.
	// This way we can test how a program reacts to being loaded at different paths
	// (e.g. "/[upload-dir]/whatever" will serve "/[upload-dir]/index.html",
	// whereas "/[upload-dir]/whatever.css" will fall
	// through to the FileServer behavior below).
	case !strings.HasSuffix(r.URL.Path, "/") && path.Ext(cleanPath) == "" && r.Method == "GET":
		// upload dir + /index.html
		f, err := http.Dir(s.BaseDir).Open("/" + strings.Split(strings.TrimPrefix(cleanPath, "/"), "/")[0] + "/index.html")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			panic(err)
		}
		http.ServeContent(w, r, "/index.html", fi.ModTime(), f)
		return

	default:

	}

	// fall through to static file server on BaseDir
	http.FileServer(http.Dir(s.BaseDir)).ServeHTTP(w, r)
	return

}
