package fixedPoint

// Exported fixed-point types
type UFix64 raw64
type Fix64 raw64

type UFix128 raw128
type Fix128 raw128

// Rounding modes
type RoundingMode int

const (
	// The Div, Mul, and FMD functions support four rounding modes:
	//    RoundTowardZero: Returns the closest representable fixed-point value that has a magnitude
	//      less than or equal to the magnitude of the real result, effectively truncating the
	//      fractional part. e.g. 5e-8 / 2 = 2e-8, -5e-8 / 2 = -2e-8
	//    RoundAwayFromZero: Returns the closest representable fixed-point value that has a magnitude
	//      greater than or equal to the magnitude of the real result, effectively rounding up
	//      any fractional part. e.g. 5e-8 / 2 = 3e-8, -5e-8 / 2 = -3e-8
	//    RoundNearestHalfAway: Returns the closest representable fixed-point value to the real result,
	//      which could be larger (rounded up) or smaller (rounded down) depending on if the
	//      unrepresentable portion is greater than or less than one half the difference between two
	//      available values. If two representable values are equally close, the value will be rounded
	//      away from zero. e.g. 7e-8 / 2 = 4e-8, 5e-8 / 2 = 3e-8
	//    RoundNearestHalfEven: Returns the closest representable fixed-point value to the real result,
	//      which could be larger (rounded up) or smaller (rounded down) depending on if the
	//      unrepresentable portion is greater than or less than one half the difference between two
	//      available values. If two representable values are equally close, the value with an even
	//      digit in the smallest decimal place will be chosen. e.g. 7e-8 / 2 = 4e-8, 5e-8 / 2 = 2e-8
	//
	// Note that for ALL rounding modes when using signed inputs, the absolute value of the result
	// will be the same regardless of the sign of the inputs.
	//
	// In other words, for all rounding modes: abs(x / y) == abs(-x / y) == abs(x / -y) == abs(-x / -y)
	RoundTowardZero RoundingMode = iota
	RoundAwayFromZero
	RoundNearestHalfAway
	RoundNearestHalfEven

	RoundTruncate = RoundTowardZero
	RoundDown     = RoundTowardZero
	RoundUp       = RoundAwayFromZero
	RoundHalfUp   = RoundNearestHalfAway
	RoundHalfEven = RoundNearestHalfEven
)

// Internal types
type raw64 uint64
type raw128 struct {
	Hi raw64
	Lo raw64
}
