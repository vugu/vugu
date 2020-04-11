package devutil

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// FileServer is similar to http.FileServer but has some options and behavior differences more useful for Vugu programs.
// The following rules apply when serving http responses:
//
// If the path is a directory but does not end with a slash it is redirected to be with a slash.
//
// If the path is a directory and ends with a slash then if it contains an index.html file that is served.
//
// If the path is a directory and ends with a slash and has no index.html, if listings are enabled a listing will be returned.
//
// If the path does not exist but exists when .html is appended to it then that file is served.
//
// For anything else the handler for the not-found case is called, or if not set then a 404.html will be searched for and if
// that's not present http.NotFound is called.
//
// Directory listings are disabled by default due to security concerns but can be enabled with SetListings.
type FileServer struct {
	fsys            http.FileSystem
	listings        bool         // do we show directory listings
	notFoundHandler http.Handler // call when not found
}

// NewFileServer returns a FileServer instance.
// Before using you must set FileSystem to serve from by calling SetFileSystem or SetDir.
func NewFileServer() *FileServer {
	return &FileServer{}
}

// SetFileSystem sets the FileSystem to use when serving files.
func (fs *FileServer) SetFileSystem(fsys http.FileSystem) *FileServer {
	fs.fsys = fsys
	return fs
}

// SetDir is short for SetFileSystem(http.Dir(dir))
func (fs *FileServer) SetDir(dir string) *FileServer {
	return fs.SetFileSystem(http.Dir(dir))
}

// SetListings enables or disables automatic directory listings when a directory is indicated in the URL path.
func (fs *FileServer) SetListings(v bool) *FileServer {
	fs.listings = v
	return fs
}

// SetNotFoundHandler sets the handle used when no applicable file can be found.
func (fs *FileServer) SetNotFoundHandler(h http.Handler) *FileServer {
	fs.notFoundHandler = h
	return fs
}

func (fs *FileServer) serveNotFound(w http.ResponseWriter, r *http.Request) {

	// notFoundHandler takes precedence
	if fs.notFoundHandler != nil {
		fs.notFoundHandler.ServeHTTP(w, r)
		return
	}

	// check for 404.html
	{
		f, err := fs.fsys.Open("/404.html")
		if err != nil {
			goto defNotFound
		}
		defer f.Close()
		st, err := f.Stat()
		if err != nil {
			goto defNotFound
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(404)
		http.ServeContent(w, r, r.URL.Path, st.ModTime(), f)
		return
	}

defNotFound:
	// otherwise fall back to http.NotFound
	http.NotFound(w, r)
}

// ServeHTTP implements http.Handler with the appropriate behavior.
func (fs *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// NOTE: much of this borrowed and adapted from https://golang.org/src/net/http/fs.go

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	const indexPage = "/index.html"

	// redirect .../index.html to .../
	// can't use Redirect() because that would make the path absolute,
	// which would be a problem running under StripPrefix
	if strings.HasSuffix(r.URL.Path, indexPage) {
		localRedirect(w, r, "./")
		return
	}

	name := path.Clean("/" + r.URL.Path)

	f, err := fs.fsys.Open(name)
	if err != nil {

		// try again with .html
		f2, err2 := fs.fsys.Open(name + ".html")
		if err2 == nil {
			f = f2
		} else {

			msg, code := toHTTPError(err)
			if code == 404 {
				fs.serveNotFound(w, r)
				return
			}
			http.Error(w, msg, code)
			return
		}

	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	// redirect to canonical path: / at end of directory url
	// r.URL.Path always begins with /
	url := r.URL.Path
	if d.IsDir() {
		if url[len(url)-1] != '/' {
			localRedirect(w, r, path.Base(url)+"/")
			return
		}
	} else {
		if url[len(url)-1] == '/' {
			localRedirect(w, r, "../"+path.Base(url))
			return
		}
	}

	if d.IsDir() {

		url := r.URL.Path
		// redirect if the directory name doesn't end in a slash
		if url == "" || url[len(url)-1] != '/' {
			localRedirect(w, r, path.Base(url)+"/")
			return
		}

		// use contents of index.html for directory, if present
		index := strings.TrimSuffix(name, "/") + indexPage
		ff, err := fs.fsys.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		} else {
			// no index.html found for directory
			if !fs.listings {
				fs.serveNotFound(w, r)
				return
			}
		}
	}

	// Still a directory? (we didn't find an index.html file)
	if fs.listings && d.IsDir() {
		if checkIfModifiedSince(r, d.ModTime()) == condFalse {
			writeNotModified(w)
			return
		}
		setLastModified(w, d.ModTime())
		dirList(w, r, f)
		return
	}

	// serveContent will check modification time
	// sizeFunc := func() (int64, error) { return d.Size(), nil }
	// serveContent(w, r, d.Name(), d.ModTime(), sizeFunc, f)

	// log.Printf("about to serve: f=%#v, d=%#v", f, d)

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

// condResult is the result of an HTTP request precondition check.
// See https://tools.ietf.org/html/rfc7232 section 3.
type condResult int

const (
	condNone condResult = iota
	condTrue
	condFalse
)

func checkIfModifiedSince(r *http.Request, modtime time.Time) condResult {
	if r.Method != "GET" && r.Method != "HEAD" {
		return condNone
	}
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" || isZeroTime(modtime) {
		return condNone
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return condNone
	}
	// The Last-Modified header truncates sub-second precision so
	// the modtime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if modtime.Before(t) || modtime.Equal(t) {
		return condFalse
	}
	return condTrue
}

func writeNotModified(w http.ResponseWriter) {
	// RFC 7232 section 4.1:
	// a sender SHOULD NOT generate representation metadata other than the
	// above listed fields unless said metadata exists for the purpose of
	// guiding cache updates (e.g., Last-Modified might be useful if the
	// response does not have an ETag field).
	h := w.Header()
	delete(h, "Content-Type")
	delete(h, "Content-Length")
	if h.Get("Etag") != "" {
		delete(h, "Last-Modified")
	}
	w.WriteHeader(http.StatusNotModified)
}

func setLastModified(w http.ResponseWriter, modtime time.Time) {
	if !isZeroTime(modtime) {
		w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	}
}

var unixEpochTime = time.Unix(0, 0)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

func dirList(w http.ResponseWriter, r *http.Request, f http.File) {
	dirs, err := f.Readdir(-1)
	if err != nil {
		log.Print(r, "http: error reading directory: %v", err)
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: name}
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), htmlReplacer.Replace(name))
	}
	fmt.Fprintf(w, "</pre>\n")
}

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",

	`"`, "&#34;",

	"'", "&#39;",
)

// ----------------------------------
// old notes:

// contentFunc     func(fs http.FileSystem, name string) (modtime time.Time, content io.ReadSeeker, err error) // can handle various request path transformations

// SetContentFunc assigns the function that will
// func (fs *FileServer) SetContentFunc(f func(fs http.FileSystem, name string) (modtime time.Time, content ReadSeekCloser, err error)) {

// }

// // DefaultContentFunc serves files directly from a the filesystem with the following additional logic:
// // If the path is a directory but does not end with a slash it is redirected to be with a slash.
// // If the path is a directory and ends with a slash then if it contains an index.html file that is served.
// // If the path does not exist but exists when .html is appended to it then that file is served.
// // For anything else the error returned from fs.Open is returned.
// func DefaultContentFunc(fs http.FileSystem, name string) (modtime time.Time, content ReadSeekCloser, err error) {

// }

// // DefaultListingContentFunc is like DefaultContentFunc but with directory listings enabled.
// func DefaultListingContentFunc(fs http.FileSystem, name string) (modtime time.Time, content ReadSeekCloser, err error) {
// }

// // ReadSeekCloser has Read, Seek and Close methods.
// type ReadSeekCloser interface {
// 	io.Reader
// 	io.Seeker
// 	io.Closer
// }

// what about /anything mapping to index page
// (seems like an option to me - maybe need some func to map this stuff
// plus convenience methods for common cases)

// pick a sensible default
