package obidefault

var __compress__ = false

func CompressOutput() bool {
	return __compress__
}

func SetCompressOutput(b bool) {
	__compress__ = b
}

func CompressOutputPtr() *bool {
	return &__compress__
}
