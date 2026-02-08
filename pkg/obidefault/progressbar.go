package obidefault

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
