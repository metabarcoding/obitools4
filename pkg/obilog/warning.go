package obilog

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"github.com/sirupsen/logrus"
)

func Warnf(format string, args ...interface{}) {
	if !obidefault.SilentWarning() {
		logrus.Warnf(format, args...)
	}
}
