package obifp

type FPUint[T Uint64 | Uint128 | Uint256] interface {
	Zero() T
	Set64(v uint64) T

	IsZero() bool
	LeftShift(n uint) T
	RightShift(n uint) T

	Add(v T) T
	Sub(v T) T
	Mul(v T) T
	//Div(v T) T

	And(v T) T
	Or(v T) T
	Xor(v T) T
	Not() T

	LessThan(v T) bool
	LessThanOrEqual(v T) bool
	GreaterThan(v T) bool
	GreaterThanOrEqual(v T) bool

	AsUint64() uint64

	Uint64 | Uint128 | Uint256
}

func ZeroUint[T FPUint[T]]() T {
	return *new(T)
}

func OneUint[T FPUint[T]]() T {
	return ZeroUint[T]().Set64(1)
}

func From64[T FPUint[T]](v uint64) T {
	return ZeroUint[T]().Set64(v)
}
