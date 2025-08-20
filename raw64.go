package fixedPoint

import (
	"math/bits"
)

var raw64Zero = raw64(0)

// This file contains wrapper functions for raw64 to provide all of the basic functionality that
// are required for the Fix64 and UFix64 types. Many of these functions are just wrappers around
// the math/bits package functions, but they are provided here to ensure that the Fix64 and UFix64 types
// can be used in a way that is consistent with Fix128 and UFix128 types.
//
// The basic operations are:
// - Addition
// - Subtraction
// - Multiplication
// - Division
// - Comparison (less than, equal to, etc.)
// - Shifting (left, right, unsigned, signed)
// - Zero and negative checks
//
// NOTE: If you check the file https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssagen/intrinsics.go
// you will see that the Go compiler replaces a lot of the bits.* functions with single CPU
// instructions on hardware that supports it. Add64, Sub64, Mul64, and LeadingZeros64 are supported
// by intrinsics on AMD64 and ARM64 architectures, while Div64 is only support on AMD64, and uses a compiled
// fallback on ARM.

func add64(a, b raw64, c uint64) (raw64, uint64) {
	// Use bits.Add64 to add two raw64 values and return the sum and carry.
	sum, carry := bits.Add64(uint64(a), uint64(b), c)
	return raw64(sum), carry
}

func sub64(a, b raw64, c uint64) (raw64, uint64) {
	// Use bits.Sub64 to subtract two raw64 values and return the difference and borrow.
	diff, borrow := bits.Sub64(uint64(a), uint64(b), c)
	return raw64(diff), borrow
}

func mul64(a, b raw64) (raw64, raw64) {
	// Use bits.Mul64 to multiply two raw64 values and return the high and low parts of the product.
	hi64, lo64 := bits.Mul64(uint64(a), uint64(b))
	return raw64(hi64), raw64(lo64)
}

func div64(a, b, y raw64) (raw64, raw64) {
	// Use bits.Div64 to divide two raw64 values and return the quotient and remainder.
	q64, r64 := bits.Div64(uint64(a), uint64(b), uint64(y))
	return raw64(q64), raw64(r64)
}

func mod64(a, b raw64) raw64 {
	// Compute the modulus of two raw64 values, treating them as unsigned integers.
	return raw64(uint64(a) % uint64(b))
}

func neg64(a raw64) raw64 {
	// Negate a raw64 value, treating it as a signed integer.
	return raw64(-int64(a))
}

func ushouldRound64(q, r, b raw64, round RoundingMode) bool {
	switch round {
	case RoundTowardZero:
		return false // Always truncate towards zero, no rounding.
	case RoundAwayFromZero:
		return r != 0 // Round away from zero, so if there's any remainder, round up.
	case RoundNearestHalfAway, RoundNearestHalfEven:
		// Determing if a particular remainder results in rounding isn't as simple
		// as just checking if r >= b/2, because dividing b by two *loses precision*.
		// A more accurate solution would be to multiply the remainder by 2 and compare
		// it to b, but that can overflow if the remainder is large.
		//
		// However, we KNOW the remainder is less than b, and we know that b fits in 64 bits;
		// if the remainder were so large that multiplying it by 2 would overflow,
		// then it must also be larger than half b. So, we first check to see if it WOULD
		// overflow when doubled (in which case it is definitely larger than b/2),
		// and otherwise we can safely double it and compare it to b.
		if uint64(r) > 0x7fffffffffffffff {
			// If r is larger than half the maximum value of a uint64, we clearly need to round up.
			return true
		}

		doubleR := uint64(r) * 2

		if doubleR > uint64(b) {
			// The remainder is strictly larger than half b, so we round up.
			return true
		} else if doubleR < uint64(b) {
			// The remainder is strictly smaller than half b, so we round down.
			return false
		} else {
			// If doubleR == b, we have to round away from zero for RoundNearestHalfAway,
			// and round to the nearest even number for RoundNearestHalfEven.
			if round == RoundNearestHalfAway {
				return true
			} else {
				return q&1 == 1
			}
		}
	default:
		panic("unsupported rounding mode")
	}
}

func leadingZeroBits64(a raw64) uint64 {
	// Count the number of leading zero bits in a raw64 value.
	// This is equivalent to bits.LeadingZeros64
	return uint64(bits.LeadingZeros64(uint64(a)))
}

func isZero64(a raw64) bool {
	// Check if a raw64 value is zero.
	return a == 0
}

func isIota64(a raw64) bool {
	// Check if a raw64 value is the iota value.
	return a == 1
}

func isNegIota64(a raw64) bool {
	return int64(a) == -1
}

func isNeg64(a raw64) bool {
	// Check if a raw64 value is negative when interpreted as a signed integer.
	return int64(a) < 0
}

func ult64(a, b raw64) bool {
	// Check if a is less than b, treating them as unsigned integers.
	return a < b
}

func slt64(a, b raw64) bool {
	// Check if a is less than b, treating them as signed integers.
	return int64(a) < int64(b)
}

func isEqual64(a, b raw64) bool {
	// Check if two raw64 values are equal.
	return a == b
}

func shiftLeft64(a raw64, shift uint64) raw64 {
	// Shift a raw64 value left by the specified number of bits.
	return a << shift
}

func ushiftRight64(a raw64, shift uint64) raw64 {
	// Shift right by a number of bits, treating it as an unsigned integer.
	return a >> shift
}

func sshiftRight64(a raw64, shift uint64) raw64 {
	// Shift right by a number of bits, treating it as an signed integer.
	return raw64(int64(a) >> shift)
}
