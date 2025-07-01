package fixedPoint

var raw128Zero = raw128{0, 0}

// This file contains methods for raw64 to provide all of the basic functionality that
// is required for the Fix128 and UFix128 types. All of the functions in this file have
// direct analogues in the raw64.go file, but they operate on 128-bit values instead
// of 64-bit values, and – in some cases – are much more complex because of it.
//
// The basic operations are:
// - Addition
// - Subtraction
// - Multiplication
// - Division
// - Comparison (less than, equal to, etc.)
// - Shifting (left, right, unsigned, signed)
// - Zero and negative checks

func add128(a, b raw128, carry uint64) (sum raw128, carryOut uint64) {
	sum.Lo, carry = add64(a.Lo, b.Lo, carry)
	sum.Hi, carryOut = add64(a.Hi, b.Hi, carry)
	return
}

func sub128(a, b raw128, borrow uint64) (diff raw128, borrowOut uint64) {
	diff.Lo, borrow = sub64(a.Lo, b.Lo, borrow)
	diff.Hi, borrowOut = sub64(a.Hi, b.Hi, borrow)
	return
}

// A utility function to perform 128x128 multiplication with a 256-bit result.
func mul128(a, b raw128) (hi, lo raw128) {

	// If either operand fits into 64 bits, we can use a simpler multiplication.
	// This also handles the case where one of the operands is zero.
	if a.Hi == 0 {
		return mul128By64(b, a.Lo)
	} else if b.Hi == 0 {
		return mul128By64(a, b.Lo)
	}

	// Observe that:
	//   a = aH•B + aL and b = bH•B + bL (where B = 2^64)
	//   a * b = (aH * bH) * B^2 + ((aH * bL) + (aL * bH)) * B + (aL * bL)
	//
	// Note that we DO NOT use Karatsuba multiplication here, because we have
	// access to efficient 64-bit multiplication, and the "Karatusba product"
	// operates on sums that could overflow 64 bits and require edge-case handling.

	// u is aH * bH
	// v is (aH * bL) + (aL * bH)
	// w is aL * bL
	var u, v1, v2 raw128
	var wHi raw64
	u.Hi, u.Lo = mul64(a.Hi, b.Hi)
	v1.Hi, v1.Lo = mul64(a.Hi, b.Lo)
	v2.Hi, v2.Lo = mul64(a.Lo, b.Hi)
	v, vCarry := add128(v1, v2, 0)
	wHi, lo.Lo = mul64(a.Lo, b.Lo)

	// The lowest word of the result (lo.Lo) was directly set when we computed w above

	// We now sum up lo.Hi, which is the low part of v plus the high part of w
	var midCarry, hiCarry uint64
	lo.Hi, midCarry = add64(v.Lo, wHi, 0)

	// The hi.Lo is the sum of the low part of u with the high part of v plus any carry
	// from the previous sum.
	hi.Lo, hiCarry = add64(u.Lo, v.Hi, midCarry)

	// hi.Hi is the high part of u plus any carry from the previous sum (and any carry from
	// computing v).
	hi.Hi, _ = add64(u.Hi, raw64(vCarry), hiCarry)

	return
}

func div128(hi, lo, y raw128) (quo raw128, rem raw128) {
	if isZero128(y) {
		panic("div128: division by zero")
	}

	// Special case: denominator fits in 64 bits
	if y.Hi == 0 {
		// If the denominator fits in 64 bits, we know that the EITHER the numerator
		// fits in 192 bits (hi.Hi == 0) OR the division will result in a value that overflows
		// 128 bits.
		if hi.Hi != 0 {
			panic("div128: overflow")
		}

		qHi, r := div64(hi.Lo, lo.Hi, y.Lo)
		qLo, r := div64(r, lo.Lo, y.Lo)
		return raw128{qHi, qLo}, raw128{0, r}
	}

	// We use the "divide and conquer" approach to compute the quotient of a 256-bit numerator
	// by a 128-bit denominator. It involves two calls to a 192 over 128 division algorithm,
	// ("3 by 2" division)
	qHi, rHi := div192by128(hi.Hi, hi.Lo, lo.Hi, y)
	qLo, rem := div192by128(rHi.Hi, rHi.Lo, lo.Lo, y)

	// Effectively multiple qHi by 2^64, assuming that qHi is under 2^64 to start with
	qHi.Hi = qHi.Lo
	qHi.Lo = 0

	quo, _ = add128(qHi, qLo, 0)

	return quo, rem
}

func neg128(a raw128) raw128 {
	// Negate a raw128 value
	negLo, borrow := sub64(0, a.Lo, 0)
	negHi, _ := sub64(0, a.Hi, borrow)
	return raw128{negHi, negLo}
}

func uintMul128(a raw128, b uint64) raw128 {
	// Perform integer multiplication of a raw128 value by a uint64 value, treating a as an unsigned integer.
	// Does NOT handle overflow, so only use internally where overflow can't happen.
	_, lo := mul128By64(a, raw64(b))
	return lo
}

func sintMul128(a raw128, b int64) raw128 {
	// Perform integer multiplication of a raw128 value by a int64 value, treating a as an signed integer.
	if b < 0 {
		// If b is negative negate the input and then multiply.
		a = neg128(a)
		b = -b
	}

	_, lo := mul128By64(a, raw64(b))
	return lo
}

func ushouldRound128(r, b raw128) bool {
	// Determing if a particular remainder results in rounding isn't as simple
	// as just checking if r >= b/2, because dividing b by two *loses precision*.
	// A more accurate solution would be to multiply the remainder by 2 and compare
	// it to b, but that can overflow if the remainder is large.
	//
	// However, we KNOW the remainder is less than b, and we know that b fits in 128 bits;
	// if the remainder were so large that multiplying it by 2 would overflow,
	// then it must also be larger than half b. So, we first check to see if it WOULD
	// overflow when doubled (in which case it is definitely larger than b/2),
	// and otherwise we can safely double it and compare it to b.
	if r.Hi > 0x7fffffffffffffff {
		// remainder is larger than 2^127, so it is definitely larger than b/2
		return true
	} else {
		twoR := shiftLeft128(r, 1)
		return !ult128(twoR, b)
	}
}

func sshouldRound128(r, b raw128) bool {
	// For signed types, we CAN just multiply the remainder by 2 and compare it to b;
	// any signed positive value (and remainders are always positive) can be safely doubled
	// within the space of an unsigned value.
	twoR := shiftLeft128(r, 1)
	return ult128(b, twoR)
}

func leadingZeroBits128(a raw128) uint64 {
	// Count the number of leading zero bits in a raw128 value.
	if a.Hi == 0 {
		return leadingZeroBits64(a.Lo) + 64
	} else {
		return leadingZeroBits64(a.Hi)
	}
}

func uintDiv128(a raw128, b uint64) raw128 {
	rawB := raw64(b)

	if ult64(a.Hi, rawB) {
		// If the high part of a is less than b, then we can use a single 64-bit division
		q, r := div64(a.Hi, a.Lo, rawB)
		if ushouldRound64(r, rawB) {
			// If the remainder is greater than half of b, round up.
			q++
		}

		return raw128{0, q}
	}

	qHi, r := div64(0, a.Hi, rawB)
	qLo, r := div64(r, a.Lo, rawB)

	if ushouldRound64(r, rawB) {
		// If the remainder is greater than half of b, round up.
		qLo++
		if qLo == 0 {
			// If we overflowed the low part, we need to increment the high part.
			qHi++
		}
	}

	return raw128{qHi, qLo}
}

func sintDiv128(a raw128, b int64) raw128 {
	var bUnsigned uint64

	sign := int64(1)
	if isNeg128(a) {
		a = sintMul128(a, -1)
		sign = -1
	}

	if b < 0 {
		// If b is negative, we need to adjust the sign.
		bUnsigned = uint64(-b)
		sign = -sign
	} else {
		bUnsigned = uint64(b)
	}

	// Does rounding for us
	res := uintDiv128(a, bUnsigned)

	return sintMul128(res, sign)
}

func isZero128(a raw128) bool {
	return isZero64(a.Hi) && isZero64(a.Lo)
}

func isIota128(a raw128) bool {
	// Check if a raw128 value is the iota value.
	return isZero64(a.Hi) && isIota64(a.Lo)
}

func isNegIota128(a raw128) bool {
	// Check if a raw128 value is the negative iota value.
	return a.Hi == 0xffffffffffffffff && a.Lo == 0xffffffffffffffff
}

func isNeg128(a raw128) bool {
	// Check if a raw128 value is negative.
	return isNeg64(a.Hi)
}

func ult128(a, b raw128) bool {
	if isEqual64(a.Hi, b.Hi) {
		// If the high parts are equal, compare the low parts.
		return ult64(a.Lo, b.Lo)
	} else {
		// If the high parts are not equal, compare them directly.
		return ult64(a.Hi, b.Hi)
	}
}

func slt128(a, b raw128) bool {
	if isEqual64(a.Hi, b.Hi) {
		// If the high parts are equal, compare the low parts.
		return slt64(a.Lo, b.Lo)
	} else {
		// If the high parts are not equal, compare them directly.
		return slt64(a.Hi, b.Hi)
	}
}

func isEqual128(a, b raw128) bool {
	return isEqual64(a.Hi, b.Hi) && isEqual64(a.Lo, b.Lo)
}

func shiftLeft128(a raw128, shift uint64) raw128 {
	if shift >= 64 {
		return raw128{Hi: a.Lo << (shift - 64), Lo: 0}
	}

	return raw128{Hi: (a.Hi << shift) | (a.Lo >> (64 - shift)), Lo: a.Lo << shift}
}

func ushiftRight128(a raw128, shift uint64) raw128 {
	if shift >= 64 {
		return raw128{Hi: 0, Lo: a.Hi >> (shift - 64)}
	}

	return raw128{Hi: a.Hi >> shift, Lo: (a.Lo >> shift) | (a.Hi << (64 - shift))}
}

func sshiftRight128(a raw128, shift uint64) raw128 {
	if shift >= 64 {
		// NOTE: We need to copy the sign bit into the high part
		return raw128{Hi: raw64(int64(a.Hi) >> 63), Lo: raw64(int64(a.Hi) >> (shift - 64))}
	}

	return raw128{Hi: raw64(int64(a.Hi) >> shift), Lo: raw64(int64(a.Lo)>>shift) | (a.Hi << (64 - shift))}
}

func unscaledRaw128(a uint64) raw128 {
	// Convert a uint64 value to a raw64 value without scaling.
	return raw128{0, raw64(a)}
}

// Helper functions for the multiplication and division algorithms above

// A utility function used in the 128x128 multiplication algorithm to efficiently
// handle multiplications where one of the operands fits in 64 bits.
func mul128By64(a raw128, b raw64) (hi, lo raw128) {
	// NOTE: Earlier versions of this function would try to "fast return"
	// when a or b were zero, but it actually resulted in slower code.
	// The go compiler can turn bits.Mul64 into a single instruction
	// so the branches are more expensive than the computation!

	// Perform multiplication using bits.Mul64. You can think about this as
	// long multiplication where our "base" is 2^64.
	//      aH  aL
	// x         b
	// -----------
	//       w   x
	// + y   z
	// -----------
	//   q   r   s
	// where:
	//   aH = high part of a (most significant 64 bits)
	//   aL = low part of a (least significant 64 bits)
	//   b  = the 64-bit multiplier
	//   w  = high part of b•aL
	//   x  = low part of b•aL
	//   y  = high part of b•aH
	//   z  = low part of b•aH
	//   q  = high part of the result, note that result fits in 192 bits, so this
	//        is is actually hi.Lo (the low part of the high 128 bits that we return)
	//   r  = mid part of the result, (lo.Hi in the return value)
	//   s  = low part of the result, (lo.Lo in the return value)
	//   (Please note that s == x)

	var w, z raw64
	var carry uint64
	w, lo.Lo = mul64(a.Lo, b)
	hi.Lo, z = mul64(a.Hi, b)

	lo.Hi, carry = add64(w, z, 0)

	// Can't overflow, since that would imply a 128 x 64 multiplication
	// overflowed 192 bits, which is not possible.
	hi.Lo, _ = add64(hi.Lo, raw64Zero, carry)

	return hi, lo
}

func div192by128(hi, mid, lo raw64, y raw128) (quo raw128, rem raw128) {
	var carry, borrow uint64

	// We assume this function is only ever called when y is >= 2^64 (i.e. y.Hi != 0).
	shift := leadingZeroBits64(y.Hi)

	// We take the 64 leading, non-zero bits of the denominator and shift it
	// into a uint64. We shift the top bits of the numerator the same amount
	// (filling in with bits from the middle value) and divide them to get
	// an estimate of the high 64-bits of the quotient. This estimate will either
	// be correct, or too high by one. (If it is too high, we will see a negative
	// remainder and can adjust.)
	estY := (y.Hi << shift) | (y.Lo >> (64 - shift))
	estHi := hi >> (64 - shift)
	estLo := (hi << shift) | (mid >> (64 - shift))

	quo.Hi, _ = div64(estHi, estLo, estY)

	// We multiply our estimate by the denominator and subtract it from the
	// original numerator to get an intermediate remainder. Note that if our
	// estimate is too high, this will result in a negative remainder, that
	// we'll have to adjust afterwards
	pHi, pLo := mul128By64(y, quo.Hi)

	// TODO: I think that pHi is always zero here, but I'm not 100% sure.
	if !isZero128(pHi) {
		panic("div192by128: estimate too high, pHi is non-zero")
	}

	// Subtract out the product from the top two parts (hi and mid) of the numerator
	// to get an interim result.
	interimMid, borrow := sub64(mid, pLo.Lo, 0)
	interimHi, borrow := sub64(hi, pLo.Hi, borrow)

	if borrow != 0 {
		// If we borrowed it means that our estimate was too high and the remainder is now negative.
		// We decrement the estimate by one, and add back in a copy of the denominator to correct
		// the interim remainder.
		quo.Hi--

		// Add back in another copy of the denominator to get the right interim remainder.
		interimMid, carry = add64(interimMid, y.Lo, 0)
		interimHi, carry = add64(interimHi, y.Hi, carry)

		if carry == 0 {
			panic("div192by128: adjusting interim remainder did not carry, is still negative")
		}
	}

	// The interim remainder (interimHi | interimMid | lo) is a 192-bit value but we know
	// that it's less than y << 64. The next step is to divide interim remaind by the
	// denominator to get the low word of the quotient and the final remainder.
	// It might look like we're right back where we started; we have a 192-bit numerator
	// (interimHi, interimMid, lo) and a 128-bit denominator (y), but we can use the fact that
	// we know that interim < y << 64 to predict that the result of this final division will fit
	// into 64 bits. We can shift the interim remainder down by (64 - shift), which is guaranteed
	// to fit 128 bits, and use the shifted y we used for the first estimate to get our final result
	finalHi := (interimHi << shift) | (interimMid >> (64 - shift))
	finalLo := (interimMid << shift) | (lo >> (64 - shift))

	// There is an edge case here where the COMPLETE interim remainder is less than the denominator,
	// but the truncated interim remainder (finalHi | finalLo) is equal to the truncated denominator (estY << 64).
	// If we find ourselves in this case, we assume at the low part of the quotient is 0xffffffffffffffff
	// and compute the remainder as (interimRemainder - y << 64 + y).
	var needsAdjustment bool
	if finalHi >= estY {
		if finalHi > estY {
			panic("div192by128: finalHi should never be greater than estY, only equal")
		}
		quo.Lo = 0xffffffffffffffff
		rem.Lo, carry = add64(lo, y.Lo, 0)
		interimMid, _ = add64(interimMid, y.Hi, carry)
		rem.Hi, _ = sub64(interimMid, y.Lo, 0)

		needsAdjustment = false
	} else {
		quo.Lo, _ = div64(finalHi, finalLo, estY)

		// Now we just need to compute the final remainder
		pHi, pLo = mul128By64(y, quo.Lo)

		// NOTE: The final of the three subtractions should always result in zero, but we still do it to
		// see if our estimate was too high, and set the borroow flag.
		rem.Lo, borrow = sub64(lo, pLo.Lo, 0)
		rem.Hi, borrow = sub64(interimMid, pLo.Hi, borrow)
		_, borrow = sub64(interimHi, pHi.Lo, borrow)

		needsAdjustment = borrow != 0
	}

	for needsAdjustment {
		// As above, our estimate could be too high, if we borrowed in that final subtraction
		// our quotiont is too high, so we need to decrement it by 1.
		quo.Lo--

		// Add a copy of the denominator to get the right remainder.
		var carry uint64
		rem.Lo, carry = add64(rem.Lo, y.Lo, 0)
		rem.Hi, carry = add64(rem.Hi, y.Hi, carry)

		if carry != 0 {
			// If the carry bit is set on the final addition, it means that we flipped
			// from a negative remainder to a positive one, so we can stop adjusting.
			needsAdjustment = false
		}
	}

	return
}
