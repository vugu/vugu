package vugufmt

import (
	"fmt"
	"strconv"
	"strings"
)

// FmtError is a formatting error.
type FmtError struct {
	Msg      string
	FileName string
	Line     int
	Column   int
}

func (e FmtError) Error() string {
	return fmt.Sprintf("%s:%v:%v: %v", e.FileName, e.Line, e.Column, e.Msg)
}

// fromGoFmt reads stdErr output from gofmt and parses it all out (if able)
func fromGoFmt(msg string) *FmtError {
	splitUp := strings.SplitN(msg, ":", 4)

	if len(splitUp) != 4 {
		return &FmtError{
			Msg: msg,
		}
	}

	line, err := strconv.Atoi(splitUp[1])
	if err != nil {
		return &FmtError{
			Msg: msg,
		}
	}

	column, err := strconv.Atoi(splitUp[2])
	if err != nil {
		return &FmtError{
			Msg: msg,
		}
	}

	return &FmtError{
		Msg:      strings.TrimSpace(splitUp[3]),
		FileName: splitUp[0],
		Line:     line,
		Column:   column,
	}
}
