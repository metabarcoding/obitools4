package obiutils

import (
	"path"
	"path/filepath"
	"strings"
)

// RemoveAllExt removes all file extensions from the given path.
//
// Parameters:
// - p: the path to remove file extensions from (string).
//
// Returns:
// - The path without any file extensions (string).
func RemoveAllExt(p string) string {

	for ext := path.Ext(p); len(ext) > 0; ext = path.Ext(p) {
		p = strings.TrimSuffix(p, ext)
	}

	return p

}

func Basename(path string) string {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)

	// Keep removing extensions until there are no more
	for ext != "" {
		filename = strings.TrimSuffix(filename, ext)
		ext = filepath.Ext(filename)
	}

	return filename
}
