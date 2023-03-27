package obiutils

import (
	"path"
	"strings"
)

func RemoveAllExt(p string) string {
  
	for ext := path.Ext(p); len(ext) > 0; ext = path.Ext(p) {
	  p = strings.TrimSuffix(p, ext)
	}
  
	return p
  
  } 