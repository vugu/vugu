package htmlx

import "io"

type writer interface {
	io.Writer
	io.ByteWriter
	WriteString(string) (int, error)
}
