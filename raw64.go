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

func uintMul64(a raw64, b uint64) raw64 {
	// Perform integer multiplication of a raw64 value by a uint64 value, treating a as an unsigned integer.
	// Does NOT handle overflow, so only use internally where overflow can't happen.
	return raw64(uint64(a) * b)
}

func sintMul64(a raw64, b int64) raw64 {
	// Perform integer multiplication of a raw64 value by a uint64 value, treating a as an signed integer.
	// Does NOT handle overflow, so only use internally where overflow can't happen.
	return raw64(int64(a) * b)
}

func ushouldRound64(r, b raw64) bool {
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
	return uint64(r) > 0x7fffffffffffffff || uint64(r)*2 >= uint64(b)
}

func sshouldRound64(r, b raw64) bool {
	// For signed types, we CAN just multiply the remainder by 2 and compare it to b;
	// any signed positive value (and remainders are always positive) can be safely doubled
	// within the space of an unsigned value.
	return uint64(r)*2 >= uint64(b)
}

func leadingZeroBits64(a raw64) uint64 {
	// Count the number of leading zero bits in a raw64 value.
	// This is equivalent to bits.LeadingZeros64
	return uint64(bits.LeadingZeros64(uint64(a)))
}

func uintDiv64(a raw64, b uint64) raw64 {
	// Perform integer division of a raw64 value by a uint64 value, treating a as an unsigned integer.
	// Rounds half up, doesn't check for division by zero or underflow.
	q := uint64(a) / b
	r := uint64(a) % b

	if ushouldRound64(raw64(r), raw64(b)) {
		// If the remainder is greater than half of b, round up.
		q++
	}

	return raw64(q)
}

func sintDiv64(a raw64, b int64) raw64 {
	// Perform integer division of a raw64 value by an int64 value, treating a as a signed integer.
	// Rounds half up, doesn't check for division by zero or underflow.
	var aUnsigned, bUnsigned uint64

	sign := int64(1)
	if isNeg64(a) {
		// If a is negative, we need to adjust the sign.
		aUnsigned = uint64(-a)
		sign = -1
	} else {
		aUnsigned = uint64(a)
	}

	if b < 0 {
		// If b is negative, we need to adjust the sign.
		bUnsigned = uint64(-b)
		sign = -sign
	} else {
		bUnsigned = uint64(b)
	}

	q := aUnsigned / bUnsigned
	r := aUnsigned % bUnsigned

	if sshouldRound64(raw64(r), raw64(b)) {
		// If the remainder is greater than half of b, round up.
		q++
	}

	return raw64(int64(q) * sign)
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

func unscaledRaw64(a uint64) raw64 {
	// Convert a uint64 value to a raw64 value without scaling.
	return raw64(a)
}
