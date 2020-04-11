package devutil

import (
	"io"
	"net/http"
	"time"
)

// raw file system but with `a` -> `a.html` and no directory listings

// FileServer is similar to http.FileServer but has some options and behavior differences more useful for Vugu programs.
// Directory listings are disabled by default (due to security concerns and it accidentally being left on in production).
type FileServer struct {
	fs          http.FileSystem
	listings    bool                                                                                        // true if directory listings are enabled
	contentFunc func(fs http.FileSystem, name string) (modtime time.Time, content io.ReadSeeker, err error) // can handle various request path transformations
}

// TODO: SetFileSystem, SetDir

func NewFileServer() *FileServer {
	return &FileServer{}
}

// what about /anything mapping to index page
// (seems like an option to me - maybe need some func to map this stuff
// plus convenience methods for common cases)

// pick a sensible default
