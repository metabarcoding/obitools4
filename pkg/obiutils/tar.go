package obiutils

import (
	"archive/tar"
	"fmt"
)

func TarFileReader(file *Reader, path string) (*tar.Reader, error) {
	tarfile := tar.NewReader(file)
	header, err := tarfile.Next()

	for err == nil {
		if header.Name == path {
			return tarfile, nil
		}
		header, err = tarfile.Next()
	}

	return nil, fmt.Errorf("file not found: %s", path)
}
