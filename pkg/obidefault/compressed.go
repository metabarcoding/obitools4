package obidefault

var __compressed__ = false

func CompressOutput() bool {
	return __compressed__
}

func SetCompressOutput(b bool) {
	__compressed__ = b
}

func CompressedPtr() *bool {
	return &__compressed__
}
