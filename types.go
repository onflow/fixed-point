package fixedPoint

// Exported fixed-point types
type UFix64 raw64
type Fix64 raw64

type UFix128 raw128
type Fix128 raw128

// Internal types
type raw64 uint64
type raw128 struct {
	Hi raw64
	Lo raw64
}
