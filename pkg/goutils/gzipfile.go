package goutils

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
)

type Wfile struct {
	compressed bool
	close      bool
	out        io.WriteCloser
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

	return &Wfile{
		compressed: compressed,
		close:      true,
		out:        fi,
		gf:         gf,
		fw:         fw,
	}, nil
}

func CompressStream(out io.WriteCloser, compressed bool, close bool) (*Wfile, error) {
	var gf *gzip.Writer
	var fw *bufio.Writer

	if compressed {
		gf = gzip.NewWriter(out)
		fw = bufio.NewWriter(gf)
	} else {
		gf = nil
		fw = bufio.NewWriter(out)
	}

	return &Wfile{
		compressed: compressed,
		close:      close,
		out:        out,
		gf:         gf,
		fw:         fw,
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

	w.fw.Flush()

	if w.compressed {
		err = w.gf.Close()
	}

	var err2 error
	err2 = nil

	if w.close {
		err2 = w.out.Close()
	}

	if err == nil {
		err = err2
	}

	return err
}
