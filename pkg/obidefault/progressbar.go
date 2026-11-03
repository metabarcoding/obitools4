package obidefault

import "os"

var __no_progress_bar__ = false

func ProgressBar() bool {
	return !__no_progress_bar__
}

func NoProgressBar() bool {
	return __no_progress_bar__
}

func SetNoProgressBar(b bool) {
	__no_progress_bar__ = b
}

func NoProgressBarPtr() *bool {
	return &__no_progress_bar__
}

// ProgressBarVisible tells whether a progress bar can be displayed: it is not
// disabled by --no-progressbar, the standard error output is a terminal and
// the standard output is not piped to another command.
func ProgressBarVisible() bool {
	if !ProgressBar() {
		return false
	}

	if e, err := os.Stderr.Stat(); err != nil || (e.Mode()&os.ModeCharDevice) != os.ModeCharDevice {
		return false
	}

	if o, err := os.Stdout.Stat(); err == nil && (o.Mode()&os.ModeNamedPipe) == os.ModeNamedPipe {
		return false
	}

	return true
}
