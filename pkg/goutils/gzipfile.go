package goutils

import (
	"bufio"
	"compress/gzip"
	"os"
)

type Wfile struct {
	compressed bool
	f          *os.File
	gf         *gzip.Writer
	fw         *bufio.Writer
}

func OpenWritingFile(name string, compressed bool, append bool) (*Wfile, error) {

	flags := os.O_WRONLY | os.O_CREATE

	if append {
		flags |= os.O_APPEND
	}
	fi, err := os.OpenFile(name, flags, 0660)
	if err != nil {
		return nil, err
	}

	var gf *gzip.Writer
	var fw *bufio.Writer

	if compressed {
		gf = gzip.NewWriter(fi)
		fw = bufio.NewWriter(gf)
	} else {
		gf = nil
		fw = bufio.NewWriter(fi)
	}

	return &Wfile{compressed: compressed,
		f:  fi,
		gf: gf,
		fw: fw,
	}, nil
}

func (w *Wfile) Write(p []byte) (n int, err error) {
	return w.fw.Write(p)
}

func (w *Wfile) WriteString(s string) (n int, err error) {
	return w.fw.Write([]byte(s))
}

func (w *Wfile) Close() error {
	var err error
	err = nil

	if w.compressed {
		err = w.gf.Close()
	}

	err2 := w.f.Close()

	if err == nil {
		err = err2
	}

	return err
}
