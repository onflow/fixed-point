package fixedPoint

// Exported fixed-point types
type UFix64 uint64
type Fix64 int64

type UFix128 raw128
type Fix128 raw128

// Internal types
type raw128 struct {
	Hi uint64
	Lo uint64
}

type fix64_extra int64
type fix128_extra raw128
