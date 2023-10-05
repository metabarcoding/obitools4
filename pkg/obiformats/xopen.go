// This is an integration of the xopen package originally written by Brent Pedersen
// (https://github.com/brentp/xopen).
//
// Here it can be considered as a fork of [Wei Shen](http://shenwei.me) the version :
//
//	https://github.com/shenwei356/xopen
//
// Package xopen makes it easy to get buffered readers and writers.
// Ropen opens a (possibly gzipped) file/process/http site for buffered reading.
// Wopen opens a (possibly gzipped) file for buffered writing.
// Both will use gzip when appropriate and will user buffered IO.
package obiformats

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	"github.com/ulikunitz/xz"
)

// Level is the default compression level of gzip.
// This value will be automatically adjusted to the default value of zstd or bzip2.
var Level = gzip.DefaultCompression

// ErrNoContent means nothing in the stream/file.
var ErrNoContent = errors.New("xopen: no content")

// ErrDirNotSupported means the path is a directory.
var ErrDirNotSupported = errors.New("xopen: input is a directory")

// IsGzip returns true buffered Reader has the gzip magic.
func IsGzip(b *bufio.Reader) (bool, error) {
	return CheckBytes(b, []byte{0x1f, 0x8b})
}

// IsXz returns true buffered Reader has the xz magic.
func IsXz(b *bufio.Reader) (bool, error) {
	return CheckBytes(b, []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00})
}

// IsZst returns true buffered Reader has the zstd magic.
func IsZst(b *bufio.Reader) (bool, error) {
	return CheckBytes(b, []byte{0x28, 0xB5, 0x2f, 0xfd})
}

// IsBzip2 returns true buffered Reader has the bzip2 magic.
func IsBzip2(b *bufio.Reader) (bool, error) {
	return CheckBytes(b, []byte{0x42, 0x5a, 0x68})
}

// IsStdin checks if we are getting data from stdin.
func IsStdin() bool {
	// http://stackoverflow.com/a/26567513
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// ExpandUser expands ~/path and ~otheruser/path appropriately
func ExpandUser(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}
	var u *user.User
	var err error
	if len(path) == 1 || path[1] == '/' {
		u, err = user.Current()
	} else {
		name := strings.Split(path[1:], "/")[0]
		u, err = user.Lookup(name)
	}
	if err != nil {
		return "", err
	}
	home := u.HomeDir
	path = home + "/" + path[1:]
	return path, nil
}

// Exists checks if a local file exits
func Exists(path string) bool {
	path, perr := ExpandUser(path)
	if perr != nil {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// CheckBytes peeks at a buffered stream and checks if the first read bytes match.
func CheckBytes(b *bufio.Reader, buf []byte) (bool, error) {

	m, err := b.Peek(len(buf))
	if err != nil {
		// return false, ErrNoContent
		return false, err // EOF
	}
	for i := range buf {
		if m[i] != buf[i] {
			return false, nil
		}
	}
	return true, nil
}

// Reader is returned by Ropen
type Reader struct {
	*bufio.Reader
	rdr io.Reader
	gz  io.ReadCloser
}

// Close the associated files.
func (r *Reader) Close() error {
	var err error
	if r.gz != nil {
		err = r.gz.Close()
		if err != nil {
			return err
		}
	}
	if c, ok := r.rdr.(io.ReadCloser); ok {
		err = c.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Writer is returned by Wopen
type Writer struct {
	*bufio.Writer
	wtr *os.File
	gz  *gzip.Writer
	xw  *xz.Writer
	zw  *zstd.Encoder
	bz2 *bzip2.Writer
}

// Close the associated files.
func (w *Writer) Close() error {
	var err error
	err = w.Flush()
	if err != nil {
		return err
	}

	if w.gz != nil {
		err = w.gz.Close()
		if err != nil {
			return err
		}
	}
	if w.xw != nil {
		err = w.xw.Close()
		if err != nil {
			return err
		}
	}
	if w.zw != nil {
		err = w.zw.Close()
		if err != nil {
			return err
		}
	}
	if w.bz2 != nil {
		err = w.bz2.Close()
		if err != nil {
			return err
		}
	}
	return w.wtr.Close()
}

// Flush the writer.
func (w *Writer) Flush() error {
	var err error
	err = w.Writer.Flush()
	if err != nil {
		return err
	}

	if w.gz != nil {
		err = w.gz.Flush()
		if err != nil {
			return err
		}
	}
	if w.zw != nil {
		err = w.zw.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

var bufSize = 65536

// Buf returns a buffered reader from an io.Reader
// If f == "-", then it will attempt to read from os.Stdin.
// If the file is gzipped, it will be read as such.
func Buf(r io.Reader) (*Reader, error) {
	b := bufio.NewReaderSize(r, bufSize)
	var rd io.Reader
	var rdr io.ReadCloser

	if is, err := IsGzip(b); err != nil {
		// check BOM
		t, _, err := b.ReadRune() // no content
		if err != nil {
			return nil, ErrNoContent
		}
		if t != '\uFEFF' {
			b.UnreadRune()
		}
		return &Reader{b, r, rdr}, nil // non-gzip file with content less than 2 bytes
	} else if is {
		rdr, err = gzip.NewReader(b)
		if err != nil {
			return nil, err
		}
		b = bufio.NewReaderSize(rdr, bufSize)
	} else if is, err := IsZst(b); err != nil {
		// check BOM
		t, _, err := b.ReadRune() // no content
		if err != nil {
			return nil, ErrNoContent
		}
		if t != '\uFEFF' {
			b.UnreadRune()
		}
		return &Reader{b, r, rdr}, nil // non-gzip/zst file with content less than 4 bytes
	} else if is {
		rd, err = zstd.NewReader(b)
		if err != nil {
			return nil, err
		}
		b = bufio.NewReaderSize(rd, bufSize)
	} else if is, err := IsXz(b); err != nil {
		// check BOM
		t, _, err := b.ReadRune() // no content
		if err != nil {
			return nil, ErrNoContent
		}
		if t != '\uFEFF' {
			b.UnreadRune()
		}
		return &Reader{b, r, rdr}, nil // non-gzip/zst/xz file with content less than 6 bytes
	} else if is {
		rd, err = xz.NewReader(b)
		if err != nil {
			return nil, err
		}
		b = bufio.NewReaderSize(rd, bufSize)
	} else if is, err := IsBzip2(b); err != nil {
		// check BOM
		t, _, err := b.ReadRune() // no content
		if err != nil {
			return nil, ErrNoContent
		}
		if t != '\uFEFF' {
			b.UnreadRune()
		}
		return &Reader{b, r, rdr}, nil // non-gzip/zst/xz file with content less than 6 bytes
	} else if is {
		rd, err = bzip2.NewReader(b, &bzip2.ReaderConfig{})
		if err != nil {
			return nil, err
		}
		b = bufio.NewReaderSize(rd, bufSize)
	}

	// other files with content >= 6 bytes

	// check BOM
	t, _, err := b.ReadRune()
	if err != nil {
		return nil, ErrNoContent
	}
	if t != '\uFEFF' {
		b.UnreadRune()
	}
	return &Reader{b, r, rdr}, nil
}

// XReader returns a reader from a url string or a file.
func XReader(f string) (io.Reader, error) {
	if strings.HasPrefix(f, "http://") || strings.HasPrefix(f, "https://") {
		var rsp *http.Response
		rsp, err := http.Get(f)
		if err != nil {
			return nil, err
		}
		if rsp.StatusCode != 200 {
			return nil, fmt.Errorf("http error downloading %s. status: %s", f, rsp.Status)
		}
		rdr := rsp.Body
		return rdr, nil
	}
	f, err := ExpandUser(f)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(f)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, ErrDirNotSupported
	}

	return os.Open(f)
}

// Ropen opens a buffered reader.
func Ropen(f string) (*Reader, error) {
	var err error
	var rdr io.Reader
	if f == "-" {
		if !IsStdin() {
			return nil, errors.New("stdin not detected")
		}
		b, err := Buf(os.Stdin)
		return b, err
	} else if f[0] == '|' {
		// TODO: use csv to handle quoted file names.
		cmdStrs := strings.Split(f[1:], " ")
		var cmd *exec.Cmd
		if len(cmdStrs) == 2 {
			cmd = exec.Command(cmdStrs[0], cmdStrs[1:]...)
		} else {
			cmd = exec.Command(cmdStrs[0])
		}
		rdr, err = cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		err = cmd.Start()
		if err != nil {
			return nil, err
		}
	} else {
		rdr, err = XReader(f)
	}
	if err != nil {
		return nil, err
	}
	b, err := Buf(rdr)
	return b, err
}

// Wopen opens a buffered reader.
// If f == "-", then stdout will be used.
// If f endswith ".gz", then the output will be gzipped.
// If f endswith ".xz", then the output will be zx-compressed.
// If f endswith ".zst", then the output will be zstd-compressed.
// If f endswith ".bz2", then the output will be bzip2-compressed.
func Wopen(f string) (*Writer, error) {
	return WopenFile(f, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// WopenFile opens a buffered reader.
// If f == "-", then stdout will be used.
// If f endswith ".gz", then the output will be gzipped.
// If f endswith ".xz", then the output will be zx-compressed.
// If f endswith ".bz2", then the output will be bzip2-compressed.
func WopenFile(f string, flag int, perm os.FileMode) (*Writer, error) {
	var wtr *os.File
	if f == "-" {
		wtr = os.Stdout
	} else {
		dir := filepath.Dir(f)
		fi, err := os.Stat(dir)
		if err == nil && !fi.IsDir() {
			return nil, fmt.Errorf("can not write file into a non-directory path: %s", dir)
		}
		if os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}
		wtr, err = os.OpenFile(f, flag, perm)
		if err != nil {
			return nil, err
		}
	}

	f2 := strings.ToLower(f)
	if strings.HasSuffix(f2, ".gz") {
		gz, err := gzip.NewWriterLevel(wtr, Level)
		if err != nil {
			err = fmt.Errorf("xopen: %s", err)
		}
		return &Writer{bufio.NewWriterSize(gz, bufSize), wtr, gz, nil, nil, nil}, err
	}
	if strings.HasSuffix(f2, ".xz") {
		xw, err := xz.NewWriter(wtr)
		return &Writer{bufio.NewWriterSize(xw, bufSize), wtr, nil, xw, nil, nil}, err
	}
	if strings.HasSuffix(f2, ".zst") {
		level := Level
		if level == gzip.DefaultCompression {
			level = 2
		}
		zw, err := zstd.NewWriter(wtr, zstd.WithEncoderLevel(zstd.EncoderLevel(level)))
		if err != nil {
			err = fmt.Errorf("xopen: zstd: %s", err)
		}
		return &Writer{bufio.NewWriterSize(zw, bufSize), wtr, nil, nil, zw, nil}, err
	}
	if strings.HasSuffix(f2, ".bz2") {
		level := Level
		if level == gzip.DefaultCompression {
			level = 6
		}
		bz2, err := bzip2.NewWriter(wtr, &bzip2.WriterConfig{Level: level})
		if err != nil {
			err = fmt.Errorf("xopen: %s", err)
		}
		return &Writer{bufio.NewWriterSize(bz2, bufSize), wtr, nil, nil, nil, bz2}, err
	}
	return &Writer{bufio.NewWriterSize(wtr, bufSize), wtr, nil, nil, nil, nil}, nil
}
