package obidefault

var __silent_warning__ = false

func SilentWarning() bool {
	return __silent_warning__
}

func SilentWarningPtr() *bool {
	return &__silent_warning__
}
